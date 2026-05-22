# Backend Code Review - Golang Task API

Vai trò: Senior Backend Engineer / Code Reviewer.

Phạm vi review: toàn bộ cấu trúc backend Go trong workspace hiện tại, tập trung vào architecture, clean code, design pattern, naming, layering, scalability, error handling, validation, DTO/mapper, GORM usage, authentication/authorization và các bug tiềm ẩn.

## Executive Summary

Codebase đang ở mức "functional prototype" tiến gần tới "small production app", nhưng chưa đạt chuẩn production backend thật sự. Điểm mạnh là đã tách được handler, service, repository, entity, middleware và route. Tuy nhiên, hiện tại vẫn tồn tại một số vấn đề kiến trúc và bảo mật quan trọng:

- Authentication đã có JWT middleware, nhưng authorization theo owner/role còn thiếu.
- Một số luồng CRUD đang trust input từ client quá nhiều.
- Project flow còn duplicate query và map not found sai HTTP status.
- Một số chỗ đang bind entity trực tiếp thay vì DTO, làm tăng coupling giữa HTTP layer và domain model.
- Validation và naming chưa đồng bộ.
- Có hardcoded secret/credential ở phần auth/database.

## Severity Summary

### Critical

1. Broken ownership / IDOR: authenticated user có thể đọc hoặc sửa tài nguyên của user khác qua `ownerID` hoặc `id` client truyền lên.
2. Task creation/update trust client input quá nhiều, dễ gán assignee sai người hoặc overwrite dữ liệu không mong muốn.

### High

1. Project update/delete đang query DB hai lần cho cùng một request.
2. Not found / forbidden / invalid input bị map sai sang HTTP 500 ở nhiều handler.
3. JWT secret và DB credential hardcoded.
4. Task update / project update có nguy cơ overwrite field khi request thiếu dữ liệu.

### Medium

1. CRUD module chưa nhất quán: project module chưa có create flow rõ ràng.
2. DTO / mapper còn lẫn entity trực tiếp trong một số handler.
3. Naming convention chưa đồng nhất (`Id` vs `ID`, `ownerID` vs `ownerId`).
4. Validation layer có coverage chưa đều.

### Low

1. Một số file import/order formatting chưa chuẩn.
2. Một số mapper function hiện chưa được dùng.
3. Response message/status chưa đồng nhất giữa các module.

---

## 1) Phân tích cấu trúc thư mục

### `cmd`

Đây là entrypoint đúng chuẩn. `cmd/main.go` chỉ nên làm bootstrap: connect DB, khởi tạo dependency, wire route, start server. Hiện tại file này đang làm đúng vai trò đó.

### `internal`

Đây là boundary tốt cho code không public ra ngoài module. Cấu trúc hiện tại dùng `internal` là hợp lý cho Go backend.

### `handler`

Đây là HTTP layer. Phần lớn handler đang làm đúng: nhận request, bind JSON, gọi service, trả JSON. Tuy nhiên, một số handler vẫn đang:

- map lỗi thành status code chưa chuẩn,
- bind trực tiếp vào entity,
- hoặc trả entity trực tiếp thay vì DTO.

Nghĩa là handler đã có separation of concerns cơ bản, nhưng chưa thật clean.

### `services`

Service layer đang giữ business rules, validation và orchestration giữa repository với handler. Đây là điểm tốt. Tuy nhiên:

- một số service đang query DB nhiều lần cho cùng một request,
- một số rule authorization lẽ ra phải nằm ở service nhưng chưa có,
- một số flow chỉ mới là "data pass-through" thay vì business use case thật.

### `repositories`

Repository layer tương đối đúng mục tiêu: chỉ làm persistence. Nhưng có chỗ bị kéo vào business semantics, ví dụ `DeleteProject(project entities.Project)` buộc service phải fetch entity trước rồi mới xóa.

Với Go backend nhỏ, chấp nhận được. Nhưng để scale tốt hơn, repository nên nhận ID hoặc criteria rõ ràng, không nên ép service fetch dư thừa.

### `entities`

Entity đang dùng như model chính cho DB. Điều này ổn ở mức nhỏ. Nhưng khi entity được dùng trực tiếp ở request/response, coupling giữa HTTP và persistence model tăng mạnh.

### `dto`

DTO là hướng đúng. Bạn đã dùng request/response DTO ở một số module user/project. Tuy nhiên chưa nhất quán:

- task handler vẫn bind trực tiếp `entities.Task`,
- project module chưa có create request DTO rõ ràng,
- mapper có function dư hoặc chưa dùng.

### `mapper`

Mapper giúp giảm coupling giữa DTO và entity, đây là điểm tốt. Nhưng hiện tại có dấu hiệu mapper bị overkill ở một số chỗ, và có function có thể panic nếu dùng không cẩn thận.

### `middleware`

Middleware auth đang là điểm cộng lớn vì đã tách được xác thực token khỏi từng handler. Nhưng middleware chỉ đang giải quyết authentication, chưa cover authorization theo owner/role.

### `validation`

Có validation layer riêng là đúng hướng. Nhưng naming chưa đồng nhất, và validation chưa được áp dụng uniformly ở mọi luồng CRUD.

### `response`

Centralized response format là tốt. Nó giúp API consistency. Điểm cần cải thiện là status code mapping và uniform error shape.

### Separation of Concerns

Tổng thể: khá đúng, nhưng chưa sạch.

Điểm ổn:

- handler không chứa nhiều business logic,
- service có vai trò điều phối,
- repository truy cập DB,
- middleware xử lý auth.

Điểm chưa ổn:

- handler vẫn biết quá nhiều về model nội bộ,
- entity bị lộ lên HTTP layer ở một số nơi,
- service update/delete project đang fetch dư thừa,
- authorization chưa được đưa vào đúng layer.

### Package nào coupling quá mức

1. `handler` coupling với `entities` ở task flow.
2. `project_service` coupling với repository implementation semantics do phải fetch entity trước khi update/delete.
3. `routes` coupling với từng concrete handler/service trong wiring là chấp nhận được ở app nhỏ, nhưng nếu scale lớn nên tách bootstrap layer rõ hơn.

### Cấu trúc clean architecture tốt hơn

Nếu muốn đi theo clean architecture rõ hơn, nên tiến dần sang:

```text
internal/
  domain/
    entities/
    repository/
  usecase/
    auth/
    task/
    project/
    user/
  delivery/
    http/
      handler/
      middleware/
      dto/
      response/
      router/
  infrastructure/
    postgres/
      repository/
    config/
```

Tuy nhiên, với project hiện tại, chưa cần big-bang rewrite. Nên refactor incremental.

---

## 2) Phân tích flow request

### Flow chuẩn mong muốn

```text
Request
 -> Handler
 -> Service
 -> Repository
 -> Database
 -> Response
```

### Đánh giá hiện tại

#### Handler

Tốt ở mức cơ bản:

- nhận HTTP request,
- bind JSON,
- parse path param,
- gọi service,
- trả response.

Nhưng có một số vấn đề:

- handler map lỗi chưa đúng HTTP status,
- một số chỗ bind entity trực tiếp,
- chưa chuẩn hóa error response.

#### Service

Đây là lớp đang làm đúng vai trò nhất. Nó chứa:

- validate ID,
- validate business input,
- check tồn tại,
- hash password,
- orchestration.

Nhưng một số luồng còn thiếu business rule authorization.

#### Repository

Repository chủ yếu chỉ làm query. Điều đó tốt. Tuy nhiên:

- có duplicate query do service fetch rồi repo fetch lại,
- một số method signature chưa tối ưu cho partial update,
- delete/update nên trả rõ rows affected hoặc not found.

### Handler có đang chứa business logic không

Nhìn chung chưa nhiều. Nhưng có mùi "logic về status code / mapping error" đang nằm rải rác ở handler. Đó là business-adjacent logic, không phải core business, nhưng nên chuẩn hóa.

### Service có xử lý business rule đúng không

Ở mức cơ bản là có. Ví dụ:

- user service validate email/password,
- auth service check password hash,
- project service validate project ID.

Nhưng thiếu:

- ownership check,
- permission check,
- transaction boundary ở một số use case phức tạp.

### Repository có bị lẫn business logic không

Ít, nhưng có dấu hiệu coupling semantic:

- `DeleteProject(project entities.Project)` buộc service phải fetch entity trước.
- `UpdateProject(projectID, name, description)` lưu kiểu "patch partial" nhưng repo vẫn load entity rồi save full.

---

## 3) Kiểm tra CRUD module

### Create

#### Task create

Hiện tại `TaskHandler.CreateTask` bind trực tiếp `entities.Task` từ JSON. Đây là design yếu vì:

- client có thể truyền `assignee_id` tùy ý,
- client có thể set field ngoài dự kiến,
- coupling giữa HTTP input và persistence model quá cao.

#### Project create

Trong project module hiện tại chưa thấy create flow hoàn chỉnh. Nếu đây là chủ đích thì không sao, nhưng nếu module được coi là CRUD thì đang thiếu một phần quan trọng.

### GetByID

#### Project GetByID

`ProjectHandler.GetProjectByID` đang map lỗi chung chung thành `500`. Đây là sai semantic. Nếu project không tồn tại thì nên trả `404`.

### List

#### ListProjectsByOwner

Đây là điểm cực kỳ nhạy cảm. Route nhận `ownerID` từ path param. Nếu chỉ dựa vào auth middleware mà không kiểm tra quyền sở hữu, bất kỳ user đã đăng nhập nào cũng có thể gọi:

```text
/api/v1/projects/owner/2
```

để xem project của owner khác.

Đây là IDOR / broken ownership check.

### Update

#### Project update

Vấn đề chính:

- service fetch project trước,
- repository update lại fetch project lần nữa,
- handler lại map mọi lỗi thành 500.

Ngoài ra `UpdateProjectRequest` dùng pointer fields là hướng tốt cho partial update, nhưng service/repo chưa xử lý đầy đủ validation và permission.

### Delete

#### Project delete

Tương tự update, service fetch entity trước rồi repo fetch lại. Dư thừa query.

Quan trọng hơn: không có ownership/role check, nên bất kỳ user authenticated nào cũng có thể xóa project nếu biết ID.

### Checklist bug theo CRUD

- Validate ID: có, nhưng chưa đồng nhất naming và mapping status.
- Validate request body: chưa đầy đủ ở project/task.
- Error handling: map chưa đúng status.
- Partial update: có hướng tiếp cận, nhưng chưa sạch.
- Duplicated query: có.
- Nil pointer: có nguy cơ ở mapper project.
- Empty update: có thể accept request rỗng và trả 200.
- Race condition: không nổi bật, nhưng thiếu transaction cho use case phức tạp.
- Unique violation: user email có check trước, nhưng vẫn nên dự phòng DB error mapping.
- Not found handling: chưa đúng ở handler.
- Permission handling: thiếu nghiêm trọng.

---

## 4) Kiểm tra DTO / Mapper

### Request DTO

Điểm tốt:

- `CreateUserRequest`, `UpdateUserRequest`, `LoginRequest` là hướng đúng.
- `UpdateProjectRequest` dùng pointer field là mẫu tốt cho patch semantics.

Điểm yếu:

- task chưa có DTO rõ ràng,
- một số handler vẫn bind entity trực tiếp,
- naming folder `dto/request/project` nhưng package tên `project` dễ gây ambiguity khi codebase lớn.

### Response DTO

Điểm tốt:

- user/project đều có response DTO, tránh leak field nội bộ.

Điểm cần cải thiện:

- task response chưa tách DTO rõ ràng,
- response format chưa hoàn toàn thống nhất giữa modules.

### Mapper usage

Điểm tốt:

- giảm việc handler tự map thủ công,
- response mapping rõ.

Điểm yếu:

- `UpdateProjectRequestToEntity` hiện không được dùng,
- nếu dùng trực tiếp, function này dereference pointer không check nil và có thể panic.

### Có đang return entity trực tiếp không

Có, rõ nhất ở task flow. Đây là chỗ nên refactor sớm nhất.

### Có bị lộ internal field không

User/project response đang che bớt tốt. Nhưng task entity vẫn được trả trực tiếp, nên nếu sau này entity có field nhạy cảm hơn thì sẽ rò rỉ.

---

## 5) Kiểm tra GORM usage

### Query efficiency

`ProjectService.UpdateProject` và `DeleteProject` tạo duplicate query do service fetch entity rồi repository lại fetch lại.

### `Save` vs `Updates`

- `Save` đang dùng ở project repository.
- `Save` có thể ghi toàn bộ struct, không tối ưu cho patch/update một vài field.
- Với partial update, nên cân nhắc `Updates` hoặc `Select`/`Omit`.

### Delete flow

- delete project đang nhận entity thay vì id.
- nếu chỉ cần xóa theo id, repo nên nhận id và delete trực tiếp hoặc delete với `Where("id = ?")`.

### Transaction thiếu ở đâu

Hiện tại chưa thấy flow nào bắt buộc transaction lớn, nhưng nếu sau này project/task/user liên quan đồng thời thì sẽ cần transaction. Ví dụ:

- tạo project + tạo default task,
- update ownership + audit log,
- batch operation.

### Preload nếu cần

Chưa thấy preload liên quan relation ở project/task/user. Nếu sau này response muốn include owner/assignee/project detail thì nên preload có kiểm soát thay vì N+1 query.

### Index cần thiết

Từ migration:

- `users.email` unique,
- `projects.owner_id` có index,
- `tasks.project_id`, `tasks.assignee_id`, `tasks.status` có index.

Đây là khá tốt cho workload hiện tại.

### Foreign key design

Thiết kế FK ổn:

- task -> project,
- task -> user (assignee),
- project -> user (owner).

Điểm cần lưu ý: code layer phải enforce ownership chứ không chỉ rely vào FK.

---

## 6) Kiểm tra validation

### Naming convention

Không đồng nhất:

- `IsValidProjectId`
- `IsValidProjectOwnerId`
- `IsValidId`
- `IsValidIdUser`

Theo Go style, nên thống nhất một chuẩn.

### Duplicated validation

Một số validate ID có vẻ bị lặp ý nghĩa giữa task/user/project. Nên gom policy rõ hơn hoặc dùng helper chung.

### Validation layer placement

Placement là đúng hướng: validation nằm ở package riêng. Nhưng cần dùng xuyên suốt service thay vì chỉ một vài flow.

### Missing validation cases

- Project name có thể chỉ check `len(name) > 0`, chưa trim space.
- Project description chỉ check length, chưa check empty content semantics nếu business yêu cầu.
- Task create/update chưa thấy request DTO validation rõ ràng.
- Ownership/permission không phải validation input thuần, nhưng lại đang thiếu hoàn toàn.

---

## 7) Kiểm tra naming convention

### `IsValidProjectID` vs `IsValidProjectId`

Trong Go, acronym thường viết theo style nhất quán hơn. Nên chọn một biến thể và dùng xuyên suốt. Hiện code đang dùng lẫn lộn `ID` và `Id`.

### Method naming

- `GetTaskById` nên cân nhắc `GetTaskByID`.
- `ListProjectByOwner` nên cân nhắc `ListProjectsByOwner` để rõ là trả nhiều project.

### Interface naming

Phần repository/interface đang ổn ở mức cơ bản. Tuy nhiên, nếu scale, nên đặt interface theo hành vi use case thay vì CRUD thuần.

### DTO naming

`CreateUserRequest`, `UpdateUserRequest` là tốt.
`UpdateProjectRequest` cũng tốt.

Điểm cần chuẩn hóa là package name `user`, `project` trong dto rất ngắn và dễ đụng tên khi codebase lớn. Có thể chấp nhận, nhưng cần consistency.

### Consistency theo Go standard

Go ưu tiên ngắn gọn, rõ nghĩa, không over-abstract. Code hiện tại đã tương đối theo hướng đó, nhưng naming còn lẫn một số chỗ và import alias chưa nhất quán.

---

## 8) Kiểm tra error handling

### Custom errors

Có, ví dụ `ErrInvalidProjectID`, `ErrProjectNotFound`, `ErrInvalidUserID`, `ErrEmailAlreadyExist`. Đây là điểm tốt.

### Wrapped errors

Chưa thấy wrap error nhiều bằng `%w`. Nên bổ sung nếu muốn trace nguyên nhân tốt hơn.

### HTTP status code usage

Đây là một trong các điểm yếu lớn nhất:

- not found đang bị trả 500,
- invalid input đôi khi trả 400 đúng, nhưng chưa đồng nhất,
- permission chưa có 403.

### Internal error leakage

Một số response đang gửi `err.Error()` trực tiếp ra client. Điều này chấp nhận được cho dev/local nhưng production nên kiểm soát kỹ hơn.

### Repository error mapping

Repository đang trả raw error từ GORM. Điều này không xấu, nhưng service/handler phải map đúng hơn.

---

## 9) Kiểm tra permission / authentication

### JWT handling

Auth middleware đã verify JWT. Đây là nền tảng tốt.

### Owner permission

Thiếu nghiêm trọng.

- `ListProjectsByOwner` trust `ownerID` từ URL.
- `UpdateProject` và `DeleteProject` không kiểm tra current user có phải owner không.
- `TaskCreate` trust `assignee_id` từ client.

### Role-based access

Chưa có role-based access thật sự.

### Middleware flow

Flow middleware hợp lý ở authentication level. Nhưng chưa đủ để đảm bảo authorization.

### Security issue

Rủi ro lớn nhất là IDOR:

- user authenticated có thể truy cập tài nguyên của user khác nếu đoán được ID.

### Trust client input issue

Đây là vấn đề hiện hữu trong task/project flow. Server không nên tin `ownerID`, `assignee_id`, hay các field ownership khác từ client nếu đây là tài nguyên private.

---

## 10) Đánh giá scalability

### Điểm mạnh

- Tách layer rõ ràng.
- Dễ đọc với project nhỏ.
- Dễ onboarding cho junior.
- GORM + Gin đủ nhanh để build MVP.

### Điểm yếu

- Không có boundaries domain/usecase rõ ràng.
- Handler và entity bị coupling ở một số luồng.
- Authorization chưa có thiết kế riêng.
- Query flow đôi khi redundant.
- Không có pagination/filter chuẩn cho list endpoints.

### Khó maintain ở đâu

- Khi thêm nhiều rule authz (owner, admin, collaborator).
- Khi request/response model diverge khỏi DB entity.
- Khi task/project/user tăng số field và logic patch/update.

### Khó scale ở đâu

- List endpoints không pagination.
- Service/repository thiếu abstraction cho transaction/use case phức tạp.
- Không có audit/logging chuẩn cho security-sensitive actions.

### Module cần refactor sớm

1. Project service + handler + repository.
2. Task create/update flow.
3. Auth/JWT config management.
4. Shared error mapping / response policy.

---

## 11) Đề xuất refactor

### Code refactor

- Tạo request DTO cho task create/update.
- Không bind trực tiếp entity từ JSON.
- Normalize status code mapping tại handler layer hoặc helper chung.
- Xóa log nhạy cảm.

### Architecture refactor

- Thêm authorization use case layer hoặc policy layer.
- Tách current user accessor helper từ gin context.
- Giảm coupling giữa handler và entity.

### Naming refactor

- Chuẩn hóa `ID` / `Id`.
- Chuẩn hóa `ownerID` / `ownerId` / `userID` theo cùng style.
- Đổi function names cho rõ plurality.

### Repository refactor

- Cho update/delete nhận ID hoặc criteria rõ ràng.
- Trả `ErrRecordNotFound`/`ErrNotFound` có map rõ.
- Hạn chế duplicate fetch.

### Service refactor

- Thêm ownership check trong project/task use case.
- Tách validate business input vs validate authorization.
- Support partial update đúng semantics.

### DTO improvement

- Tạo `CreateTaskRequest`, `UpdateTaskRequest`.
- Project update DTO nên có validation tags hoặc validator layer rõ.
- Không expose entity trực tiếp.

### Permission improvement

- Lấy current user từ JWT middleware.
- Không tin owner/user ID từ client nếu không phải admin.
- Tách admin route nếu có RBAC.
- Dùng middleware/policy để enforce quyền theo resource owner.

---

## 12) Các lỗi tiềm ẩn cực kỳ quan trọng

### 1. Broken ownership check

Critical.

Các route project hiện tại cho phép user authenticated truy cập tài nguyên theo `ownerID` và `id` từ URL. Nếu không check owner hoặc role, đây là IDOR.

### 2. Task assignee spoofing

Critical.

Client có thể truyền `assignee_id` bất kỳ khi create task. Nếu business rule yêu cầu task phải thuộc current user hoặc chỉ admin mới assign người khác, hiện tại đang thiếu bảo vệ.

### 3. Duplicate query + stale design

High.

Service fetch trước, repository fetch lại. Tốn query, tăng latency, code khó đọc.

### 4. Not found bị trả 500

High.

Sai HTTP semantics, làm client khó xử lý và che mất issue thật.

### 5. Hidden panic risk trong mapper

Medium-High.

`UpdateProjectRequestToEntity` dereference pointer mà không nil-check. Hiện chưa dùng, nhưng nếu tái sử dụng sẽ dễ panic.

### 6. Overwrite field khi update

High.

Task update repo đang set thẳng tất cả field từ struct update. Nếu request thiếu field hoặc có zero value, dữ liệu cũ bị ghi đè ngoài ý muốn.

### 7. Hardcoded secret/credential

High.

JWT secret và DB connection string đang hardcode trong source.

### 8. Broken ownership on list endpoint

Critical.

`ListProjectsByOwner` nhận `ownerID` từ path param. Đây là một lỗ hổng authorization rất dễ khai thác.

---

## 13) Ví dụ code clean hơn

### Ví dụ 1: Project update nên kiểm tra owner và tránh duplicate fetch

```go
func (s *ProjectService) UpdateProject(currentUserID, projectID int, req UpdateProjectRequest) error {
    if !validation.IsValidProjectId(projectID) {
        return ErrInvalidProjectID
    }

    project, err := s.projectRepository.GetProjectByID(projectID)
    if err != nil {
        return ErrProjectNotFound
    }

    if project.OwnerID != currentUserID {
        return ErrForbidden
    }

    updates := map[string]interface{}{}
    if req.Name != nil {
        updates["name"] = *req.Name
    }
    if req.Description != nil {
        updates["description"] = *req.Description
    }

    if len(updates) == 0 {
        return ErrEmptyUpdate
    }

    return s.projectRepository.UpdateProjectFields(projectID, updates)
}
```

### Ví dụ 2: Handler map status code đúng hơn

```go
func mapProjectErrorToStatus(err error) int {
    switch {
    case errors.Is(err, services.ErrInvalidProjectID), errors.Is(err, services.ErrInvalidProjectOwnerId):
        return http.StatusBadRequest
    case errors.Is(err, services.ErrProjectNotFound):
        return http.StatusNotFound
    case errors.Is(err, services.ErrForbidden):
        return http.StatusForbidden
    default:
        return http.StatusInternalServerError
    }
}
```

### Ví dụ 3: Create task không trust client assignee_id

```go
currentUserIDAny, exists := c.Get("user_id")
if !exists {
    c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized", "missing user context"))
    return
}

currentUserID, ok := currentUserIDAny.(int)
if !ok {
    c.JSON(http.StatusInternalServerError, response.ErrorResponse("Internal error", "invalid user context"))
    return
}

task.AssigneeID = &currentUserID
```

---

## 14) So sánh với production backend thực tế

### Hiện tại giống production ở điểm nào

- Có layer rõ.
- Có JWT auth.
- Có bcrypt.
- Có validation cơ bản.
- Có response wrapper.

### Khác production ở điểm nào

- Authorization theo resource chưa có.
- Config/secrets chưa quản lý chuẩn.
- Error semantics chưa đồng nhất.
- CRUD chưa clean DTO boundary.
- Transaction, pagination, audit logging, RBAC còn thiếu.

### Kết luận thực tế

Đây là một project tốt để học kiến trúc Go backend. Nhưng nếu đưa vào production thật cho môi trường nhiều user, mình sẽ bắt buộc refactor các điểm sau trước:

1. Ownership/authorization.
2. JWT secret + DB credential.
3. DTO boundary cho task/project.
4. Error mapping và status code.
5. Duplicate query / update semantics.

---

## 15) Kết luận ngắn

Codebase có nền tảng ổn cho một backend junior/intern project, nhưng chưa đạt chuẩn production-ready. Vấn đề lớn nhất không phải syntax hay style, mà là authorization, query semantics và coupling giữa HTTP layer với model nội bộ.

Nếu refactor theo thứ tự ưu tiên sau thì hiệu quả nhất:

1. Sửa ownership/permission.
2. Tách DTO khỏi entity cho task/project.
3. Chuẩn hóa error handling/status code.
4. Dẹp duplicate query và tối ưu GORM usage.
5. Dọn naming/validation để codebase dễ scale hơn.
