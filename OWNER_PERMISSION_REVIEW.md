# Owner Permission Review Report

## 1. Tổng quan hiện trạng

Project hiện tại có sơ bộ sẵn cấu trúc authorization ở service layer. Middleware JWT đã được gắn vào protected routes, và `ProjectService` có kiểm tra `owner_id`. Tuy nhiên, phân quyền vẫn **chưa toàn diện** - một số API endpoint vẫn cho phép cross-user access:

- ✅ `ProjectService.UpdateProject` / `DeleteProject` kiểm tra owner
- ✅ `TaskService` đã bổ sung kiểm tra project owner khi create/update/delete
- ✅ `TaskService.GetAllTasks` filter tasks từ projects của owner
- ❌ `ProjectService.GetProjectByID` không kiểm tra owner
- ❌ `ProjectService.ListAllProjects` trả toàn bộ projects
- ❌ `TaskService.GetTaskById` không nhận currentUserID, không kiểm tra owner
- ❌ Repository layer không có ownership validation

## 2. Luồng phân quyền hiện tại

```
HTTP Request (với Authorization header)
    ↓
AuthMiddleware
  ├─ Validate JWT token từ Authorization header
  ├─ Extract user_id từ JWT claims
  ├─ Verify user tồn tại trong DB
  └─ Lưu current_user và user_id vào Gin context
    ↓
Handler
  ├─ Extract currentUserID từ context bằng utils.CurrentUserID()
  ├─ Truyền currentUserID xuống service
  └─ Map response
    ↓
Service
  ├─ Kiểm tra currentUserID có quyền không (đối với create/update/delete)
  ├─ Gọi repository
  └─ Trả về kết quả hoặc error
    ↓
Repository
  ├─ Query DB (TUY NHIÊN: không validate owner)
  └─ Trả về data
    ↓
Database
```

**Lưu ý:** `currentUserID` được extract từ JWT token trong middleware, sau đó:
- Set vào Gin context với key `"user_id"` (line 65 auth.go)
- Lấy ra bằng `utils.CurrentUserID(c)` ở handler
- Truyền xuống service
- **NHƯNG:** Repository không biết ownership, chỉ query theo ID nguyên thủy

## 3. Danh sách API cần kiểm tra

| API | Mục đích | Handler | Service | Repo | Đã check owner? | Rủi ro | Mức độ |
|---|---|---|---|---|---|---|---|
| GET /api/v1/projects/me | List projects của current user | ✓ lấy currentUserID | ✓ filter ownerID | ✓ WHERE owner_id | **✓ OK** | Không | - |
| GET /api/v1/projects | List tất cả projects | ✗ không filter | ✗ trả toàn bộ | ✓ Find() | **✗ FAIL** | User A thấy project của User B | **Critical** |
| GET /api/v1/projects/:id | Get project by ID | ✗ không check currentUserID | ✗ không validate owner | ✗ First(id) | **✗ FAIL** | User A lấy project ID của User B | **Critical** |
| POST /api/v1/projects | Create project | ✓ set currentUserID làm owner | (nên có) | ✓ Create | **~ Cần verify** | Nên OK nếu service check | Medium |
| PUT /api/v1/projects/:id | Update project | ✓ kiểm tra forbidden | ✓ check owner_id = currentUserID | ✗ không validate owner | **✓ Tương đối OK** | Nếu repo không kiểm tra, có thể bypass | Medium |
| DELETE /api/v1/projects/:id | Delete project | ✓ kiểm tra forbidden | ✓ check owner_id = currentUserID | ✗ không validate owner | **✓ Tương đối OK** | Nếu repo không kiểm tra, có thể bypass | Medium |
| GET /api/v1/tasks | List tasks của current user | ✓ truyền currentUserID | ✓ lấy projects của owner, filter tasks | ✓ WHERE project_id IN (...) | **✓ OK** | Không | - |
| GET /api/v1/tasks/:id | Get task by ID | **✗ không lấy currentUserID** | **✗ không validate owner** | ✗ First(id) | **✗ FAIL** | User A lấy task ID của User B | **Critical** |
| POST /api/v1/tasks | Create task | ✓ truyền currentUserID | ✓ check project.owner_id = currentUserID | ✓ Create | **✓ OK** | Không | - |
| PUT /api/v1/tasks/:id | Update task | ✓ truyền currentUserID | ✓ check project.owner_id = currentUserID | ✗ không validate owner | **✓ Tương đối OK** | Nếu repo không kiểm tra, bypass | Medium |
| DELETE /api/v1/tasks/:id | Delete task | ✓ truyền currentUserID | ✓ check project.owner_id = currentUserID | ✗ không validate owner | **✓ Tương đối OK** | Nếu repo không kiểm tra, bypass | Medium |

## 4. Lỗ hỏng bảo mật phát hiện được

### Lỗi số 1: API GET /api/v1/projects/:id cho phép xem project của người khác

**Vị trí code:**

- `ProjectHandler.GetProjectByID` ([internal/handler/project_handler.go](internal/handler/project_handler.go#L46-L55))
- `ProjectService.GetProjectByID` ([internal/services/project_service.go](internal/services/project_service.go#L36-L44))
- `ProjectRepository.GetProjectByID` ([internal/repositories/project_repository.go](internal/repositories/project_repository.go#L60-L67))

**Vấn đề:**

```go
// ProjectHandler - không lấy currentUserID
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
    projectID, err := strconv.Atoi(c.Param("id"))
    // ...
    project, err := h.service.GetProjectByID(projectID)  // ✗ không truyền currentUserID
    // ...
}

// ProjectService - không kiểm tra owner
func (s *ProjectService) GetProjectByID(projectID int) (entities.Project, error) {
    project, err := s.projectRepository.GetProjectByID(projectID)  // ✗ không validate owner
    return project, nil
}

// ProjectRepository - chỉ query ID
func (r *ProjectRepository) GetProjectByID(projectID int) (entities.Project, error) {
    err := r.db.First(&project, projectID).Error  // ✗ SELECT * FROM projects WHERE id = ?
    return project, nil
}
```

**Tình huống nguy hiểm:**

```
User A: GET /api/v1/projects/5
→ Handler trả về project ID 5 của User B (nếu User B tồn tại)
→ User A lấy được: { id: 5, name: "Dự án bí mật", owner_id: 2, ... }
→ Cross-user data leak
```

**Mức độ:** Critical

**Cách sửa:**

Handler cần lấy currentUserID và kiểm tra owner:

```go
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
    currentUserID, err := utils.CurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
        return
    }
    
    projectID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
        return
    }
    
    // ✓ Thêm currentUserID để service kiểm tra ownership
    project, err := h.service.GetProjectByID(currentUserID, projectID)
    if err != nil {
        if errors.Is(err, services.ErrForbidden) {
            c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
            return
        }
        if errors.Is(err, services.ErrProjectNotFound) {
            c.JSON(http.StatusNotFound, response.ErrorResponse("Project not found", err.Error()))
            return
        }
        c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to get project", err.Error()))
        return
    }
    c.JSON(http.StatusOK, response.SuccessResponse("Project retrieved successfully", mapper.ProjectToResponse(project)))
}
```

Service cần kiểm tra owner:

```go
func (s *ProjectService) GetProjectByID(currentUserID, projectID int) (entities.Project, error) {
    if !validation.IsValidProjectId(projectID) {
        return entities.Project{}, ErrInvalidProjectID
    }
    
    project, err := s.projectRepository.GetProjectByID(projectID)
    if err != nil {
        return entities.Project{}, ErrProjectNotFound
    }
    
    // ✓ Kiểm tra ownership
    if project.OwnerID != currentUserID {
        return entities.Project{}, ErrForbidden
    }
    
    return project, nil
}
```

---

### Lỗi số 2: API GET /api/v1/projects trả về tất cả projects của mọi user

**Vị trí code:**

- `ProjectHandler.ListAllProjects` ([internal/handler/project_handler.go](internal/handler/project_handler.go#L38-L44))
- `ProjectService.ListAllProjects` ([internal/services/project_service.go](internal/services/project_service.go#L30-L35))
- `ProjectRepository.ListAllProjects` ([internal/repositories/project_repository.go](internal/repositories/project_repository.go#L38-L46))

**Vấn đề:**

Route cấu hình hai endpoint:
- `GET /projects/me` → ListMyProjects (✓ safe, filter owner)
- `GET /projects` → ListAllProjects (✗ unsafe, trả toàn bộ)

API bị lộ dữ liệu của tất cả project trong hệ thống.

**Tình huống nguy hiểm:**

```
User A: GET /api/v1/projects
→ Trả về tất cả 100+ projects của mọi user trong hệ thống
→ User A enumeration và phân tích được toàn bộ projects
→ Có thể reverse-engineer cấu trúc công ty, dự án khác
```

**Mức độ:** Critical

**Cách sửa:**

Xóa route ListAllProjects hoặc di chuyển sang admin-only endpoint:

```go
// ✓ Thay vì trả toàn bộ projects, chỉ trả projects của current user
// Cách 1: Xóa ListAllProjects khỏi public routes
// Cách 2: Hoặc admin-only với role check

// Trong routes/project_routes.go
func (r *ProjectRoutes) SetupProjectRoutes(router *gin.RouterGroup) {
    projectGroup := router.Group("/projects")
    {
        projectGroup.GET("/me", r.projectHandler.ListMyProjects)  // ✓ OK
        // ✗ Bỏ projectGroup.GET("", r.projectHandler.ListAllProjects)
        projectGroup.GET("/:id", r.projectHandler.GetProjectByID)
        projectGroup.POST("", r.projectHandler.CreateProject)
        projectGroup.PUT("/:id", r.projectHandler.UpdateProject)
        projectGroup.DELETE("/:id", r.projectHandler.DeleteProject)
    }
}
```

---

### Lỗi số 3: API GET /api/v1/tasks/:id cho phép xem task của người khác

**Vị trí code:**

- `TaskHandler.GetTaskById` ([internal/handler/task_handler.go](internal/handler/task_handler.go#L33-L44))
- `TaskService.GetTaskById` ([internal/services/task_service.go](internal/services/task_service.go#L25-L33))
- `TaskRepository.GetTaskById` ([internal/repositories/task_repository.go](internal/repositories/task_repository.go#L42-L49))

**Vấn đề:**

```go
// TaskHandler - không lấy currentUserID
func (h *TaskHandler) GetTaskById(c *gin.Context) {
    // ✗ Không lấy currentUserID từ context
    mappedTask, err := h.service.GetTaskById(id)  // ✗ Không truyền currentUserID
}

// TaskService - không validate ownership
func (s *TaskService) GetTaskById(id int) (*entities.Task, error) {
    getTask, err := s.taskRepo.GetTaskById(id)
    // ✗ Không kiểm tra xem task có thuộc project của user không
    return getTask, nil
}
```

**Tình huống nguy hiểm:**

```
Task A thuộc Project 10 (owner = User B)
User A: GET /api/v1/tasks/123 (task A)
→ Handler trả về task A của User B
→ User A lấy được: { id: 123, project_id: 10, title: "Task bí mật", ... }
→ Cross-user data leak
```

**Mức độ:** Critical

**Cách sửa:**

Handler lấy currentUserID và truyền xuống:

```go
func (h *TaskHandler) GetTaskById(c *gin.Context) {
    // ✓ Lấy currentUserID
    currentUserID, err := utils.CurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
        return
    }
    
    idString := c.Param("id")
    id, err := strconv.Atoi(idString)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to get task", errors.New("invalid task ID").Error()))
        return
    }
    
    // ✓ Truyền currentUserID
    mappedTask, err := h.service.GetTaskById(currentUserID, id)
    if err != nil {
        if errors.Is(err, services.ErrForbidden) {
            c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
            return
        }
        c.JSON(http.StatusNotFound, response.ErrorResponse("Failed to get task", errors.New("task not found").Error()))
        return
    }
    c.JSON(http.StatusOK, response.SuccessResponse("Task found", mapper.ToTaskResponse(*mappedTask)))
}
```

Service kiểm tra ownership:

```go
// ✓ Thêm currentUserID tham số
func (s *TaskService) GetTaskById(currentUserID, id int) (*entities.Task, error) {
    if !validation.IsValidId(id) {
        return nil, errors.New("invalid id")
    }
    
    getTask, err := s.taskRepo.GetTaskById(id)
    if err != nil {
        return nil, err
    }
    
    // ✓ Kiểm tra xem task có thuộc project của user không
    project, err := s.projectRepo.GetProjectByID(getTask.ProjectID)
    if err != nil {
        return nil, errors.New("project not found")
    }
    
    if project.OwnerID != currentUserID {
        return nil, errors.New("forbidden: task does not belong to your project")
    }
    
    return getTask, nil
}
```

---

### Lỗi số 4: Repository layer không validate ownership khi update/delete

**Vị trí code:**

- `ProjectRepository.UpdateProject` ([internal/repositories/project_repository.go](internal/repositories/project_repository.go#L46-L56))
- `ProjectRepository.DeleteProject` ([internal/repositories/project_repository.go](internal/repositories/project_repository.go#L68-L70))
- `TaskRepository.UpdateTask` ([internal/repositories/task_repository.go](internal/repositories/task_repository.go#L68-L90))
- `TaskRepository.DeleteTask` ([internal/repositories/task_repository.go](internal/repositories/task_repository.go#L92-L96))

**Vấn đề:**

Mặc dù Service layer kiểm tra ownership, Repository layer không validate. Nếu ai đó gọi repository trực tiếp (hoặc qua lỗi logic ở service), có thể vô tình modify resource của người khác.

```go
// ✗ Repository không kiểm tra owner_id
func (r *ProjectRepository) UpdateProject(projectID int, name string, description string) error {
    var project entities.Project
    err := r.db.First(&project, projectID).Error  // ✗ Không filter owner
    if err != nil {
        return err
    }
    project.Name = name
    project.Description = description
    return r.db.Save(&project).Error
}

// ✗ Repository không kiểm tra ownership
func (r *TaskRepository) UpdateTask(id int, task entities.Task) (*entities.Task, error) {
    // UPDATE tasks SET title=?, ... WHERE id=?
    // ✗ Không có WHERE project_id IN (SELECT id FROM projects WHERE owner_id = ?)
}
```

**Tình huống nguy hiểm:**

Nếu service code có bug (ví dụ quên check ownership ở một case nào đó), repository vẫn sẽ execute query mà không validate.

**Mức độ:** Medium (vì service layer có check, nhưng defense-in-depth yêu cầu repository cũng check)

**Cách sửa:**

Repository nên nhận tham số `currentUserID` hoặc `owner_id` để validate:

```go
// ✓ Thêm ownership check
func (r *ProjectRepository) UpdateProject(projectID int, ownerID int, name string, description string) error {
    var project entities.Project
    // ✓ WHERE id = ? AND owner_id = ?
    err := r.db.Where("id = ? AND owner_id = ?", projectID, ownerID).First(&project).Error
    if err != nil {
        return err
    }
    project.Name = name
    project.Description = description
    return r.db.Save(&project).Error
}

// ✓ Task update với owner validation qua JOIN
func (r *TaskRepository) UpdateTask(id int, ownerID int, task entities.Task) (*entities.Task, error) {
    var existingTask entities.Task
    // ✓ JOIN projects và check owner_id
    err := r.db.
        Joins("JOIN projects ON projects.id = tasks.project_id").
        Where("tasks.id = ? AND projects.owner_id = ?", id, ownerID).
        First(&existingTask).Error
    
    if err != nil {
        return nil, err
    }
    
    existingTask.Title = task.Title
    existingTask.Description = task.Description
    existingTask.Status = task.Status
    existingTask.AssigneeID = task.AssigneeID
    
    err = r.db.Save(&existingTask).Error
    if err != nil {
        return nil, err
    }
    
    return &existingTask, nil
}
```

---

## 5. Review từng tầng

### 5.1 Middleware (AuthMiddleware)

**Kiểm tra:**

| Điểm | Kết quả | Ghi chú |
|---|---|---|
| JWT validate đúng | ✓ OK | Dùng `utils.ValidateAccessToken()`, check method HS256, verify signature |
| Lấy được `user_id` từ JWT | ✓ OK | Extract từ claims["user_id"], cast float64 → int |
| Validate user tồn tại trong DB | ✓ OK | Gọi `userRepo.GetUserByID()` để verify |
| Lưu vào context Gin | ✓ OK | `c.Set("current_user", user)` và `c.Set("user_id", user.ID)` |
| Xử lý lỗi 401 đúng | ✓ OK | Trả 401 cho missing header, invalid format, invalid token, user not found |

**Kết luận:** Middleware tốt, không có vấn đề.

---

### 5.2 Handler

| Handler | Điểm | Kết quả | Vấn đề |
|---|---|---|---|
| **ProjectHandler.ListMyProjects** | Lấy currentUserID | ✓ | Không |
| | Truyền xuống service | ✓ | Không |
| | HTTP status | ✓ | 200 OK |
| **ProjectHandler.ListAllProjects** | Lấy currentUserID | ✗ | Không cần, nhưng endpoint sai (trả toàn bộ) |
| | Kiểm tra owner | ✗ | **CRITICAL** |
| | HTTP status | ✓ | 200 OK |
| **ProjectHandler.GetProjectByID** | Lấy currentUserID | ✗ | **CRITICAL** |
| | Truyền xuống service | ✗ | **CRITICAL** |
| | Kiểm tra forbidden | ✗ | **CRITICAL** |
| | HTTP status | ✓ | 200 OK (but wrong data) |
| **ProjectHandler.UpdateProject** | Lấy currentUserID | ✓ | OK |
| | Truyền xuống service | ✓ | OK |
| | Handle ErrForbidden | ✓ | 403 Forbidden |
| | HTTP status | ✓ | 200 OK |
| **ProjectHandler.DeleteProject** | Lấy currentUserID | ✓ | OK |
| | Truyền xuống service | ✓ | OK |
| | Handle ErrForbidden | ✓ | 403 Forbidden |
| | HTTP status | ✓ | 200 OK |
| **TaskHandler.GetAllTasks** | Lấy currentUserID | ✓ | OK |
| | Truyền xuống service | ✓ | OK |
| | HTTP status | ✓ | 200 OK |
| **TaskHandler.GetTaskById** | Lấy currentUserID | ✗ | **CRITICAL** |
| | Truyền xuống service | ✗ | **CRITICAL** |
| | HTTP status | ✗ | Trả 200 cho unauthorized access |
| **TaskHandler.CreateTask** | Lấy currentUserID | ✓ | OK |
| | Truyền xuống service | ✓ | OK |
| | Kiểm tra unauthorized | ✓ | 400 (nên 403) |
| | HTTP status | ✓ | 201 Created |
| **TaskHandler.UpdateTask** | Lấy currentUserID | ✓ | OK |
| | Truyền xuống service | ✓ | OK |
| | HTTP status | ✓ | 200 OK |
| **TaskHandler.DeleteTask** | Lấy currentUserID | ✓ | OK |
| | Truyền xuống service | ✓ | OK |
| | HTTP status | ✗ | 404 (hardcode, nên use constant) |

---

### 5.3 Service

| Service | Điểm | Kết quả | Vấn đề |
|---|---|---|---|
| **ProjectService.ListProjectsByOwner** | Nhận currentUserID | ✓ | OK (tên hàm là ownerID nhưng logic đúng) |
| | Validate input | ✓ | IsValidProjectOwnerId |
| | Filter owner | ✓ | WHERE owner_id = ? |
| | Tách biệt lỗi | ✓ | Tra về error |
| **ProjectService.ListAllProjects** | Nhận currentUserID | ✗ | Không kiểm tra, trả toàn bộ |
| | Kiểm tra owner | ✗ | **CRITICAL** |
| **ProjectService.GetProjectByID** | Nhận currentUserID | ✗ | **CRITICAL** - không nhận tham số |
| | Kiểm tra owner | ✗ | **CRITICAL** |
| | Tách biệt ErrForbidden vs ErrNotFound | ✗ | Không thể phân biệt |
| **ProjectService.UpdateProject** | Nhận currentUserID | ✓ | OK |
| | Kiểm tra ownership | ✓ | project.OwnerID != currentUserID |
| | Return ErrForbidden | ✓ | OK |
| **ProjectService.DeleteProject** | Nhận currentUserID | ✓ | OK |
| | Kiểm tra ownership | ✓ | project.OwnerID != currentUserID |
| | Return ErrForbidden | ✓ | OK |
| **TaskService.GetAllTasks** | Nhận currentUserID | ✓ | OK |
| | Lấy projects của owner | ✓ | ListProjectByOwner |
| | Lấy tasks từ projects | ✓ | GetTaskListByProjectID |
| | Đúng filter | ✓ | OK |
| **TaskService.GetTaskById** | Nhận currentUserID | ✗ | **CRITICAL** - không nhận tham số |
| | Kiểm tra owner | ✗ | **CRITICAL** |
| **TaskService.CreateTask** | Nhận currentUserID | ✓ | OK |
| | Kiểm tra project ownership | ✓ | project.OwnerID != currentUserID |
| | Return unauthorized | ✓ | "unauthorized to create task" |
| **TaskService.UpdateTask** | Nhận currentUserID | ✓ | OK |
| | Kiểm tra project ownership | ✓ | project.OwnerID != currentUserID |
| | Return unauthorized | ✓ | "unauthorized to update this task" |
| **TaskService.DeleteTask** | Nhận currentUserID | ✓ | OK |
| | Kiểm tra project ownership | ✓ | project.OwnerID != currentUserID |
| | Return unauthorized | ✓ | "unauthorized to delete this task" |

---

### 5.4 Repository

| Repository | Query | Ownership check | Vấn đề |
|---|---|---|---|
| **ProjectRepository.ListProjectByOwner** | `WHERE owner_id = ?` | ✓ | OK |
| **ProjectRepository.ListAllProjects** | `Find(&projects)` | ✗ | **CRITICAL** - lấy toàn bộ |
| **ProjectRepository.GetProjectByID** | `First(&project, id)` | ✗ | **CRITICAL** - không filter owner |
| **ProjectRepository.UpdateProject** | `WHERE id = ?` sau `First()` | ✗ | **MEDIUM** - nên thêm `AND owner_id = ?` |
| **ProjectRepository.DeleteProject** | `Delete(&project)` | ✗ | **MEDIUM** - nên kiểm tra owner trước |
| **TaskRepository.GetAllTasks** | `Find(&tasks)` | ✗ | **CRITICAL** - lấy toàn bộ |
| **TaskRepository.GetTaskListByProjectID** | `WHERE project_id = ?` | ✓ | OK |
| **TaskRepository.GetTaskById** | `First(&task, id)` | ✗ | **CRITICAL** - không check ownership |
| **TaskRepository.CreateTask** | `Create(&task)` | ✓ | OK (Service kiểm tra project owner) |
| **TaskRepository.UpdateTask** | `WHERE id = ?` | ✗ | **MEDIUM** - nên `JOIN projects WHERE id = ? AND projects.owner_id = ?` |
| **TaskRepository.DeleteTask** | `Delete(&task, id)` | ✗ | **MEDIUM** - nên kiểm tra ownership |

---

## 6. Code nên sửa

### 6.1 Get all tasks của current user

**Hiện tại:** ✓ Đã OK

```go
func (s *TaskService) GetAllTasks(currentUserID int) ([]entities.Task, error) {
    // ✓ Lấy projects của current user
    projects, err := s.projectRepo.ListProjectByOwner(currentUserID)
    if err != nil {
        return nil, err
    }
    
    var allTasks []entities.Task
    for _, project := range projects {
        // ✓ Lấy tasks từ mỗi project
        tasks, err := s.taskRepo.GetTaskListByProjectID(project.ID)
        if err != nil {
            return nil, err
        }
        allTasks = append(allTasks, tasks...)
    }
    return allTasks, nil
}
```

---

### 6.2 Get task by ID (CRITICAL FIX NEEDED)

**Vấn đề hiện tại:** Không kiểm tra owner

**Fix:**

Handler:
```go
func (h *TaskHandler) GetTaskById(c *gin.Context) {
    // ✓ Bắt buộc lấy currentUserID
    currentUserID, err := utils.CurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
        return
    }
    
    idString := c.Param("id")
    id, err := strconv.Atoi(idString)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", errors.New("invalid task ID").Error()))
        return
    }
    
    // ✓ Truyền currentUserID
    mappedTask, err := h.service.GetTaskById(currentUserID, id)
    if err != nil {
        if errors.Is(err, services.ErrForbidden) {
            c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
            return
        }
        c.JSON(http.StatusNotFound, response.ErrorResponse("Task not found", err.Error()))
        return
    }
    c.JSON(http.StatusOK, response.SuccessResponse("Task found", mapper.ToTaskResponse(*mappedTask)))
}
```

Service:
```go
// ✓ Thêm currentUserID tham số
func (s *TaskService) GetTaskById(currentUserID, id int) (*entities.Task, error) {
    if !validation.IsValidId(id) {
        return nil, errors.New("invalid id")
    }
    
    getTask, err := s.taskRepo.GetTaskById(id)
    if err != nil {
        return nil, errors.New("task not found")
    }
    
    // ✓ Kiểm tra xem task có thuộc project của user không
    project, err := s.projectRepo.GetProjectByID(getTask.ProjectID)
    if err != nil {
        return nil, errors.New("project not found")
    }
    
    if project.OwnerID != currentUserID {
        return nil, errors.New("forbidden")
    }
    
    return getTask, nil
}
```

---

### 6.3 Update task

**Hiện tại:** ✓ Đã kiểm tra ownership ở service

```go
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    currentID, err := utils.CurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
        return
    }
    
    taskID := c.Param("id")
    id, err := strconv.Atoi(taskID)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", err.Error()))
        return
    }
    
    var req TaskRequestDTO.UpdateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request body", err.Error()))
        return
    }
    
    updateTask := mapper.UpdateTaskRequestToTaskEntity(req)
    
    // ✓ Service kiểm tra ownership
    updatedTask, err := h.service.UpdateTask(id, currentID, updateTask)
    if err != nil {
        if errors.Is(err, errors.New("unauthorized to update this task")) {
            c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
            return
        }
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Failed to update task", err.Error()))
        return
    }
    
    c.JSON(http.StatusOK, response.SuccessResponse("Task updated successfully", updatedTask))
}
```

**Nâng cấp Repository với ownership check:**

```go
func (r *TaskRepository) UpdateTask(id int, ownerID int, task entities.Task) (*entities.Task, error) {
    var existingTask entities.Task
    
    // ✓ Join projects để check owner_id
    err := r.db.
        Joins("JOIN projects ON projects.id = tasks.project_id").
        Where("tasks.id = ? AND projects.owner_id = ?", id, ownerID).
        First(&existingTask).Error
    
    if err != nil {
        return nil, err
    }
    
    existingTask.Title = task.Title
    existingTask.Description = task.Description
    existingTask.Status = task.Status
    existingTask.AssigneeID = task.AssigneeID
    
    err = r.db.Save(&existingTask).Error
    if err != nil {
        return nil, err
    }
    
    return &existingTask, nil
}
```

---

### 6.4 Delete task

**Hiện tại:** ✓ Đã kiểm tra ownership ở service

```go
func (h *TaskHandler) DeleteTask(c *gin.Context) {
    taskID := c.Param("id")
    id, err := strconv.Atoi(taskID)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid task ID", err.Error()))
        return
    }
    
    currentUserID, err := utils.CurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
        return
    }
    
    // ✓ Service kiểm tra ownership
    err = h.service.DeleteTask(id, currentUserID)
    if err != nil {
        if errors.Is(err, errors.New("unauthorized to delete this task")) {
            c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
            return
        }
        c.JSON(http.StatusNotFound, response.ErrorResponse("Task not found", err.Error()))
        return
    }
    
    c.JSON(http.StatusOK, response.SuccessResponse("Task deleted successfully", nil))
}
```

---

### 6.5 Get project by ID (CRITICAL FIX NEEDED)

**Vấn đề hiện tại:** Không kiểm tra owner

**Fix:**

Handler:
```go
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
    // ✓ Bắt buộc lấy currentUserID
    currentUserID, err := utils.CurrentUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", err.Error()))
        return
    }
    
    projectID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid project ID", err.Error()))
        return
    }
    
    // ✓ Truyền currentUserID
    project, err := h.service.GetProjectByID(currentUserID, projectID)
    if err != nil {
        if errors.Is(err, services.ErrForbidden) {
            c.JSON(http.StatusForbidden, response.ErrorResponse("Forbidden", err.Error()))
            return
        }
        if errors.Is(err, services.ErrProjectNotFound) {
            c.JSON(http.StatusNotFound, response.ErrorResponse("Project not found", err.Error()))
            return
        }
        c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to get project", err.Error()))
        return
    }
    
    c.JSON(http.StatusOK, response.SuccessResponse("Project retrieved successfully", mapper.ProjectToResponse(project)))
}
```

Service:
```go
// ✓ Thêm currentUserID tham số
func (s *ProjectService) GetProjectByID(currentUserID, projectID int) (entities.Project, error) {
    if !validation.IsValidProjectId(projectID) {
        return entities.Project{}, ErrInvalidProjectID
    }
    
    project, err := s.projectRepository.GetProjectByID(projectID)
    if err != nil {
        return entities.Project{}, ErrProjectNotFound
    }
    
    // ✓ Kiểm tra ownership
    if project.OwnerID != currentUserID {
        return entities.Project{}, ErrForbidden
    }
    
    return project, nil
}
```

---

### 6.6 Update project

**Hiện tại:** ✓ Đã kiểm tra ownership ở service

Không cần thay đổi handler, nhưng nâng cấp repository:

```go
func (r *ProjectRepository) UpdateProject(projectID int, ownerID int, name string, description string) error {
    var project entities.Project
    
    // ✓ Thêm ownership check
    err := r.db.Where("id = ? AND owner_id = ?", projectID, ownerID).First(&project).Error
    if err != nil {
        return err
    }
    
    project.Name = name
    project.Description = description
    
    return r.db.Save(&project).Error
}
```

---

### 6.7 Delete project

**Hiện tại:** ✓ Đã kiểm tra ownership ở service

Không cần thay đổi handler, nhưng nâng cấp repository:

```go
func (r *ProjectRepository) DeleteProject(project entities.Project) error {
    // ✓ Nên thêm WHERE owner_id = ? để confirm ownership
    return r.db.Where("id = ? AND owner_id = ?", project.ID, project.OwnerID).Delete(&project).Error
}
```

---

## 7. Repository query đề xuất (GORM pattern an toàn)

### Pattern 1: Get project by ID with ownership check

```go
func (r *ProjectRepository) GetProjectByID(projectID int, ownerID int) (entities.Project, error) {
    var project entities.Project
    // ✓ WHERE id = ? AND owner_id = ?
    err := r.db.Where("id = ? AND owner_id = ?", projectID, ownerID).First(&project).Error
    return project, err
}
```

### Pattern 2: Get task by ID with ownership check via JOIN

```go
func (r *TaskRepository) GetTaskById(id int, ownerID int) (*entities.Task, error) {
    var task entities.Task
    // ✓ JOIN projects ON projects.id = tasks.project_id
    // ✓ WHERE tasks.id = ? AND projects.owner_id = ?
    err := r.db.
        Joins("JOIN projects ON projects.id = tasks.project_id").
        Where("tasks.id = ? AND projects.owner_id = ?", id, ownerID).
        First(&task).Error
    
    if err != nil {
        return nil, err
    }
    return &task, nil
}
```

### Pattern 3: List tasks của current user

```go
func (r *TaskRepository) GetTasksByOwner(ownerID int) ([]entities.Task, error) {
    var tasks []entities.Task
    // ✓ JOIN projects, WHERE owner_id = ?, GROUP BY project
    err := r.db.
        Joins("JOIN projects ON projects.id = tasks.project_id").
        Where("projects.owner_id = ?", ownerID).
        Order("tasks.id ASC").
        Find(&tasks).Error
    
    return tasks, err
}
```

### Pattern 4: Update task with ownership check

```go
func (r *TaskRepository) UpdateTask(id int, ownerID int, updatedTask entities.Task) (*entities.Task, error) {
    var existingTask entities.Task
    
    // ✓ Bước 1: Tìm task và kiểm tra owner
    err := r.db.
        Joins("JOIN projects ON projects.id = tasks.project_id").
        Where("tasks.id = ? AND projects.owner_id = ?", id, ownerID).
        First(&existingTask).Error
    
    if err != nil {
        return nil, err
    }
    
    // ✓ Bước 2: Update
    existingTask.Title = updatedTask.Title
    existingTask.Description = updatedTask.Description
    existingTask.Status = updatedTask.Status
    existingTask.AssigneeID = updatedTask.AssigneeID
    
    err = r.db.Save(&existingTask).Error
    if err != nil {
        return nil, err
    }
    
    return &existingTask, nil
}
```

### Pattern 5: Delete task with ownership check

```go
func (r *TaskRepository) DeleteTask(id int, ownerID int) error {
    // ✓ WHERE id IN (SELECT id FROM tasks JOIN projects WHERE project_id AND owner_id = ?)
    // Hoặc dùng subquery
    return r.db.
        Where("id IN (SELECT t.id FROM tasks t JOIN projects p ON p.id = t.project_id WHERE t.id = ? AND p.owner_id = ?)", id, ownerID).
        Delete(&entities.Task{}).Error
}
```

---

## Tóm tắt các lỗi theo mức độ

| Mức độ | Lỗi | Ảnh hưởng | Priority |
|---|---|---|---|
| **Critical** | GET /projects/:id không check owner | User A xem project của User B | **P0 - Sửa ngay** |
| **Critical** | GET /projects trả toàn bộ projects | Enumeration data leak | **P0 - Sửa ngay** |
| **Critical** | GET /tasks/:id không check owner | User A xem task của User B | **P0 - Sửa ngay** |
| **Medium** | Repository không validate owner khi update/delete | Phòng chống bug ở service | **P1 - Sửa tiếp** |
| **Medium** | HTTP status code không nhất quán | 404 vs 403 | **P2** |

---

## Checklist sửa

- [ ] ProjectHandler.GetProjectByID - thêm currentUserID check
- [ ] ProjectService.GetProjectByID - validate ownership
- [ ] ProjectHandler.ListAllProjects - xóa hoặc di chuyển sang admin-only
- [ ] TaskHandler.GetTaskById - lấy currentUserID, truyền xuống service
- [ ] TaskService.GetTaskById - validate ownership
- [ ] ProjectRepository - add ownerID check cho UPDATE/DELETE
- [ ] TaskRepository - add ownerID check cho UPDATE/DELETE
- [ ] Update interface contracts cho repo methods
- [ ] Unit tests cho ownership validation
- [ ] Integration tests cho cross-user scenarios
