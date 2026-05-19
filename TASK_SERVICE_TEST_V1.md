# Hướng dẫn test cho Task Service

## Mục đích
Tài liệu này mô tả cách chạy và debug các unit test cho `task` service trong dự án.

## Vị trí file test
- File test chính: internal/services/task_service_test.go
- Mã nguồn service: internal/services/task_service.go
- Mock data (nếu dùng): internal/data/mock_data.go

## Yêu cầu trước khi chạy
- Cài đặt Go (>=1.20).
- Trong GOPATH/Go module: đã chạy `go mod tidy`.
- Biến môi trường cần thiết (nếu test tương tác DB) phải được set hoặc dùng mock.

## Các lệnh phổ biến
- Chạy tất cả test trong dự án:
```bash
go test ./...
```
- Chạy chỉ các test trong package `internal/services`:
```bash
go test ./internal/services -v
```
- Chạy một test cụ thể (ví dụ `TestCreateTask`):
```bash
go test ./internal/services -run TestCreateTask -v
```
- Sinh báo cáo coverage và mở HTML:
```bash
go test ./internal/services -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Case test chính cần quan tâm
- Tạo task thành công (`CreateTask`)
- Lấy task theo ID (`GetTaskByID`)
- Cập nhật task (`UpdateTask`)
- Xóa task (`DeleteTask`)
- Xử lý validation (trường bắt buộc, định dạng sai)
- Xử lý lỗi từ repository (DB lỗi, timeout)
- Trường hợp dữ liệu rỗng / không tìm thấy

## Cách viết/mocking
- Nếu test phụ thuộc DB, ưu tiên dùng mock repository (`repositories`) hoặc mock data trong `internal/data/mock_data.go`.
- Sử dụng table-driven tests để mở rộng nhanh các trường hợp.
- Mỗi test nên độc lập, reset/mocking trạng thái giữa các test.

## Debug và troubleshooting
- Thêm `-v` để xem output chi tiết.
- Dùng `t.Log` trong test để in các thông tin hỗ trợ.
- Nếu test chạy khác môi trường local, kiểm tra biến môi trường và config trong `internal/config`.

## Ghi chú
- Cập nhật tài liệu này khi thêm test mới hoặc khi flow service thay đổi.

---
