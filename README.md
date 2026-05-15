# Task Management API - Go Project

Một ứng dụng API quản lý công việc được xây dựng bằng **Go** và **Gin Framework**.

## 📋 Mục lục

- [Tổng quan về dự án](#tổng-quan-về-dự-án)
- [Cấu trúc thư mục](#cấu-trúc-thư-mục)
- [Công nghệ sử dụng](#công-nghệ-sử-dụng)
- [Cài đặt và chạy](#cài-đặt-và-chạy)
- [API Endpoints](#api-endpoints)
- [Cấu trúc kiến trúc](#cấu-trúc-kiến-trúc)
- [Các thành phần chính](#các-thành-phần-chính)

---

## 🎯 Tổng quan về dự án

**Task Management API** là một ứng dụng RESTful API cho phép quản lý các công việc. Ứng dụng cung cấp các chức năng:

- ✅ Xem tất cả công việc
- ✅ Xem chi tiết công việc theo ID
- ✅ Tạo công việc mới
- ✅ Cập nhật công việc
- ✅ Xóa công việc

---

## 📁 Cấu trúc thư mục

```
Week1_Golang/
├── cmd/
│   └── main.go                    # Điểm vào ứng dụng
├── internal/
│   ├── config/                    # Cấu hình ứng dụng
│   ├── data/
│   │   └── mock_data.go          # Dữ liệu mô phỏng
│   ├── entities/
│   │   └── task.go               # Cấu trúc Task
│   ├── handler/
│   │   └── task_handler.go       # Xử lý request HTTP
│   ├── middleware/
│   │   ├── logger.go             # Middleware ghi log
│   │   ├── recovery.go           # Middleware xử lý lỗi
│   │   └── request_id.go         # Middleware tạo request ID
│   ├── repositories/
│   │   └── task_repository.go    # Truy cập dữ liệu
│   ├── response/
│   │   └── response.go           # Định dạng response
│   ├── routes/
│   │   └── routes.go             # Định tuyến API
│   ├── services/
│   │   └── task_service.go       # Logic nghiệp vụ
│   └── validation/
│       └── task_validation.go    # Xác thực dữ liệu
├── go.mod                         # Module definition
└── README.md                      # Tài liệu này
```

---

## 🛠️ Công nghệ sử dụng

| Công nghệ | Phiên bản | Mô tả |
|-----------|----------|--------|
| **Go** | 1.26.3 | Ngôn ngữ lập trình |
| **Gin** | 1.12.0 | Web framework |
| **MongoDB Driver** | 2.5.0 | Trình điều khiển MongoDB (có thể sử dụng trong tương lai) |
| **Validator** | 10.30.1 | Xác thực dữ liệu |

---

## 🚀 Cài đặt và chạy

### Yêu cầu
- Go 1.26.3 trở lên
- Git (tuỳ chọn)

### Bước 1: Clone hoặc tải dự án
```bash
cd Week1_Golang
```

### Bước 2: Cài đặt dependencies
```bash
go mod download
```

### Bước 3: Chạy ứng dụng
```bash
go run cmd/main.go
```

Ứng dụng sẽ chạy trên `http://localhost:8080`

### Bước 4: Kiểm tra ứng dụng
```bash
curl http://localhost:8080/api/v1/tasks
```

---

## 📡 API Endpoints

### 1. Lấy tất cả công việc
**GET** `/api/v1/tasks`

```bash
curl http://localhost:8080/api/v1/tasks
```

**Response:**
```json
{
  "message": "get all tasks successfully",
  "data": [
    {
      "id": 1,
      "title": "Task 1",
      "description": "Description 1",
      "status": "TODO",
      "assignee": "User 1"
    }
  ]
}
```

---

### 2. Lấy công việc theo ID
**GET** `/api/v1/tasks/:id`

```bash
curl http://localhost:8080/api/v1/tasks/1
```

**Response:**
```json
{
  "message": "Task found",
  "data": {
    "id": 1,
    "title": "Task 1",
    "description": "Description 1",
    "status": "TODO",
    "assignee": "User 1"
  }
}
```

---

### 3. Tạo công việc mới (Chưa kích hoạt)
**POST** `/api/v1/tasks`

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Task",
    "description": "Task Description",
    "status": "TODO",
    "assignee": "User Name"
  }'
```

**Status hợp lệ:** `TODO`, `IN_PROGRESS`, `DONE`

---

### 4. Cập nhật công việc (Chưa kích hoạt)
**PUT** `/api/v1/tasks/:id`

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Task",
    "description": "Updated Description",
    "status": "IN_PROGRESS",
    "assignee": "Updated User"
  }'
```

---

### 5. Xóa công việc (Chưa kích hoạt)
**DELETE** `/api/v1/tasks/:id`

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1
```

---

## 🏗️ Cấu trúc kiến trúc

Ứng dụng sử dụng mô hình **Clean Architecture** với các layer sau:

```
Request HTTP
    ↓
Routes (routes.go)
    ↓
Handler (task_handler.go) - Xử lý request/response
    ↓
Service (task_service.go) - Logic nghiệp vụ
    ↓
Repository (task_repository.go) - Truy cập dữ liệu
    ↓
Data (mock_data.go) - Dữ liệu
```

---

## 🔧 Các thành phần chính

### 1. **Entities** (`internal/entities/task.go`)
Định nghĩa cấu trúc dữ liệu Task:

```go
type Task struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Status      string `json:"status"`
    Assignee    string `json:"assignee"`
}
```

### 2. **Handler** (`internal/handler/task_handler.go`)
Xử lý các HTTP request và trả về response:

- `GetAllTasks()` - Lấy tất cả tasks
- `GetTaskById()` - Lấy task theo ID
- `CreateTask()` - Tạo task mới
- `UpdateTask()` - Cập nhật task
- `DeleteTask()` - Xóa task

### 3. **Service** (`internal/services/task_service.go`)
Chứa logic nghiệp vụ và xác thực:

- Kiểm tra status hợp lệ: `TODO`, `IN_PROGRESS`, `DONE`
- Kiểm tra ID hợp lệ
- Gọi Repository để thực hiện thao tác dữ liệu

### 4. **Repository** (`internal/repositories/task_repository.go`)
Truy cập và quản lý dữ liệu:

- `GetAllTasks()` - Lấy tất cả
- `GetTaskById()` - Lấy theo ID
- `CreateTask()` - Thêm task mới
- `UpdateTask()` - Cập nhật task
- `DeleteTask()` - Xóa task

### 5. **Routes** (`internal/routes/routes.go`)
Định tuyến API endpoints (v1):

```go
GET    /api/v1/tasks       - Lấy tất cả
GET    /api/v1/tasks/:id   - Lấy theo ID
POST   /api/v1/tasks       - Tạo task (Chưa kích hoạt)
PUT    /api/v1/tasks/:id   - Cập nhật task (Chưa kích hoạt)
DELETE /api/v1/tasks/:id   - Xóa task (Chưa kích hoạt)
```

### 6. **Middleware** (`internal/middleware/`)
- `logger.go` - Ghi log request
- `recovery.go` - Xử lý panic
- `request_id.go` - Tạo ID duy nhất cho mỗi request

### 7. **Response** (`internal/response/response.go`)
Định dạng response thống nhất cho API.

---

## 📝 Ghi chú

- Hiện tại API sử dụng **mock data** lưu trong bộ nhớ
- Các endpoint POST, PUT, DELETE được comment ra trong `routes.go`
- Để sử dụng MongoDB, cần bỏ comment và cấu hình kết nối

---

## 🤝 Đóng góp

Nếu bạn muốn cải thiện dự án, vui lòng:

1. Fork dự án
2. Tạo branch cho feature của bạn
3. Commit thay đổi
4. Push lên branch
5. Tạo Pull Request

---

## 📄 License

Dự án này được cung cấp cho mục đích học tập.

---

## 📧 Liên hệ

Nếu có câu hỏi hoặc gợi ý, vui lòng liên hệ qua email hoặc tạo issue trên GitHub.

---

**Ngày cập nhật:** 15 Tháng 5, 2026
