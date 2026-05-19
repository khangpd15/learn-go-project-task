# Hướng dẫn chạy và kiểm thử (Tiếng Việt)

Mục đích: Hướng dẫn cách chạy unit test cho project REST API (Golang) này, bao gồm test handlers sử dụng `httptest` và mock repository.

Yêu cầu:
- Đã cài Go (phiên bản hỗ trợ modules). Kiểm tra bằng `go version`.
- Đảm bảo đang ở thư mục gốc của dự án (nơi có `go.mod`).

Các lệnh cơ bản (PowerShell/CMD):
```powershell
cd d:\Week1_Golang
go test ./... -v
```

Chạy chỉ các test trong thư mục `test`:
```powershell
go test ./test -v
```

Chạy một test cụ thể:
```powershell
go test ./... -run TestGetAllTasks_AuthAndRole -v
```

Các tùy chọn hữu ích:
- `-race` để dò race condition: `go test ./... -race -v`
- Coverage: `go test ./... -coverprofile=coverage.out` và `go tool cover -html=coverage.out`

Ghi chú về test handlers trong `test/handlers_test.go`:
- Tests sử dụng `httptest` để tạo request/response giả và `gin` ở chế độ `TestMode`.
- Middleware `AuthMiddleware` yêu cầu header `User-ID` cho các route `/api/v1/*`.
- Các test đã mock `UserRepositoryInterface` và `TaskRepositoryInterface` bằng struct triển khai interface.

Ví dụ nhanh: test tạo task thành công
1. Chạy:
```powershell
go test ./test -run TestCreateTask_BadRequest_Forbidden_Success -v
```
2. Kết quả mong đợi: test kiểm tra cả trường hợp `403 Forbidden` cho role không có quyền, `400 Bad Request` khi body không hợp lệ, và `201 Created` cho tạo thành công.

Xử lý sự cố:
- Nếu nhận lỗi `command not found: go` → cài Go từ https://go.dev/dl/ và thêm `go` vào `PATH`.
- Nếu test báo lỗi import (không tìm thấy package `task_api/internal/...`) → chạy `go env GOPATH` và đảm bảo bạn đang ở thư mục gốc dự án trước khi chạy `go test`.
- Chạy `go mod tidy` nếu thiếu dependency.

Muốn mình chạy tests trên máy này không (nếu `go` sẵn), hay bổ sung assertions kiểm tra nội dung JSON trả về? Hãy cho biết bạn muốn chi tiết thêm phần nào.
