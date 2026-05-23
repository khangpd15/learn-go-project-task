# Task Management API

REST API viết bằng Go, Gin, GORM và PostgreSQL theo hướng Clean Architecture.

## Tính năng
- Auth `register` / `login` bằng JWT
- Quản lý `users`, `projects`, `tasks`
- PostgreSQL schema qua migrations SQL

## Công nghệ
- Go 1.26.3
- Gin 1.12.0
- GORM 1.31.1
- PostgreSQL
- JWT (`github.com/golang-jwt/jwt/v5`)

## Chạy local
```bash
go mod download
go test ./...
go run cmd/main.go
```

## Cấu hình
Ứng dụng hiện dùng PostgreSQL local trong `internal/database/postgres.go` và cần `JWT_SECRET` để ký token.

Nếu chạy bằng Docker Compose:
```bash
docker compose up -d db
```

## Endpoint chính

- `GET /health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### Protected
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/:id`
- `POST /api/v1/tasks`
- `PUT /api/v1/tasks/:id`
- `DELETE /api/v1/tasks/:id`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `GET /api/v1/users/email/:email`
- `GET /api/v1/users/fullname/:fullname`
- `POST /api/v1/users`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`
- `GET /api/v1/projects/me`
- `GET /api/v1/projects`
- `GET /api/v1/projects/:id`
- `PUT /api/v1/projects/:id`
- `DELETE /api/v1/projects/:id`

## Auth mẫu
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"fullname":"Test User","email":"test@example.com","password":"Password@123"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password@123"}'
```

Protected routes cần header `Authorization: Bearer <token>`.

## Test
```bash
go test ./...
```

## Ghi chú
- Các route protected dùng middleware JWT.
- Status task hợp lệ: `TODO`, `IN_PROGRESS`, `DONE`.

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
- Các endpoint `POST`, `PUT`, `DELETE` đã bật trong `internal/routes/routes.go`
- `GET /api/v1/tasks` chấp nhận vai trò `GUEST`, `CUSTOMER`, `ADMIN`
- `GET /api/v1/tasks/:id`, `POST`, `PUT` yêu cầu `CUSTOMER` hoặc `ADMIN`
- `DELETE /api/v1/tasks/:id` chỉ cho `ADMIN`
- Nếu muốn đổi dữ liệu khởi tạo, sửa `internal/data/mock_data.go`

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
