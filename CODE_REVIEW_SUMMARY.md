# Code Review Re-check Summary

## 1. Tổng quan
So với file review cũ, code hiện tại đã tốt hơn ở phần auth middleware và database constraints, nhưng vẫn chưa đạt mức production-ready. Clean Architecture đang đúng hướng ở mức cơ bản, migrations đã có PK/FK/UNIQUE/check constraint, và project/task service đã có ownership check cho update/delete.

Các vấn đề còn lại vẫn rõ: log lộ password/hash, `Register` của AuthService vẫn check duplicate sai thứ tự và chưa map lỗi DB đúng, `/users` vẫn quá mở, Task update bị lệch status casing, và một số handler vẫn map status code chưa chuẩn.

## 2. Những điểm đã fix
- Auth middleware đã chuyển sang JWT Bearer thay vì trust header `User-ID`. Xem [internal/middleware/auth.go](internal/middleware/auth.go#L13).
- JWT generation và validation đang có thật, không còn kiểu mock header auth. Xem [internal/utils/jwt.go](internal/utils/jwt.go).
- UserService đã dùng `ExistsByEmail` đúng cách khi tạo user và `ExistsByEmailAndIDNot` khi update. Xem [internal/services/user_service.go](internal/services/user_service.go#L50).
- Project update/delete đã có ownership check ở service và handler map 403/404 tương đối đúng. Xem [internal/services/project_service.go](internal/services/project_service.go#L57) và [internal/handler/project_handler.go](internal/handler/project_handler.go#L59).
- Task create/update/delete đã kiểm tra project tồn tại và ownership ở service. Xem [internal/services/task_service.go](internal/services/task_service.go#L65).
- Migrations đã có unique email, FK, ON DELETE CASCADE/SET NULL, status check và indexes. Xem [migrations/000001_create_users_projects_tasks.up.sql](migrations/000001_create_users_projects_tasks.up.sql) và [migrations/000002_harden_constraints_indexes.up.sql](migrations/000002_harden_constraints_indexes.up.sql).
- `internal/response.ApiResponse` đã được dùng ở phần lớn handler. Xem [internal/response/response.go](internal/response/response.go).

## 3. Những điểm chưa fix
- Critical: vẫn còn log lộ password/hash. `AuthService.Login` in email, hash, password ra stdout, và `UserHandler.CreateUser` cũng log password raw. Xem [internal/services/auth_service.go](internal/services/auth_service.go#L35) và [internal/handler/user_handler.go](internal/handler/user_handler.go#L86).
- High: `AuthService.Register` vẫn check duplicate email trước khi validate input, và đang gọi `GetUserByEmail` thay vì `ExistsByEmail`. Nếu repository trả lỗi DB khác `record not found`, code hiện tại vẫn không phân biệt rõ. Xem [internal/services/auth_service.go](internal/services/auth_service.go#L52) và [internal/repositories/user_repository.go](internal/repositories/user_repository.go#L143).
- High: `/users` endpoints vẫn chỉ được bảo vệ bằng JWT, không có self-only check hoặc role check. Với requirement chỉ có role `USER`, đây là rủi ro truy cập dữ liệu người khác. Xem [internal/routes/user_routes.go](internal/routes/user_routes.go#L20) và [internal/routes/routes.go](internal/routes/routes.go#L32).
- High: Task status update vẫn chưa thống nhất. Request binding nhận lowercase `todo`, `in_progress`, `done` nhưng validation service chỉ chấp nhận uppercase `TODO`, `IN_PROGRESS`, `DONE`. Xem [internal/dto/request/task/update_task_request.go](internal/dto/request/task/update_task_request.go#L3) và [internal/validation/task_validation.go](internal/validation/task_validation.go#L3).
- Medium: Project CRUD vẫn thiếu `CreateProject` route/service/repository. Nếu project creation là requirement thì đây là thiếu chức năng. Xem [internal/routes/project_routes.go](internal/routes/project_routes.go#L20) và [internal/services/project_service.go](internal/services/project_service.go#L39).
- Medium: `ProjectHandler.GetProjectByID` vẫn trả 500 cho mọi lỗi thay vì map not found thành 404. Xem [internal/handler/project_handler.go](internal/handler/project_handler.go#L42).
- Medium: `TaskHandler.GetTaskById` so sánh lỗi bằng `errors.New("forbidden")`, nên nhánh 403 gần như không bao giờ chạy. Xem [internal/handler/task_handler.go](internal/handler/task_handler.go#L49) và [internal/services/task_service.go](internal/services/task_service.go#L42).
- Medium: Unit tests vẫn ở dạng comment block, chưa có test thực thi thật cho AuthService/TaskService. Xem [internal/services/task_service_test.go](internal/services/task_service_test.go) và [internal/handler/task_handler_test.go](internal/handler/task_handler_test.go).
- Low/Medium: README vẫn mô tả cũ, còn nói theo kiểu `User-ID` header/mock data, không khớp code hiện tại. Xem [README.md](README.md).
- Low/Medium: Database config đang hardcode DSN và password trong source, chưa dùng env config đúng nghĩa. Xem [internal/database/postgres.go](internal/database/postgres.go#L11).

## 4. Những điểm fix sai hoặc cần tối ưu thêm
- `internal/handler/auth_handler.go::Register` đang map mọi lỗi thành 409. Cách đúng là tách 400 cho validate sai và 409 cho duplicate email. Xem [internal/handler/auth_handler.go](internal/handler/auth_handler.go#L48).
- `internal/services/auth_service.go::Register` cần đổi sang validate trước, sau đó dùng `ExistsByEmail`, và phải dừng ngay nếu repository trả lỗi DB. Hiện tại logic này chưa an toàn. Xem [internal/services/auth_service.go](internal/services/auth_service.go#L52).
- `internal/handler/task_handler.go::GetTaskById` cần so sánh bằng sentinel error hoặc `errors.Is` với biến error dùng chung, không phải `errors.New("forbidden")` tạo mới mỗi lần. Xem [internal/handler/task_handler.go](internal/handler/task_handler.go#L51).
- `internal/handler/project_handler.go::GetProjectByID` nên map `ErrProjectNotFound` sang 404 thay vì đẩy hết về 500. Xem [internal/handler/project_handler.go](internal/handler/project_handler.go#L42).
- `internal/database/postgres.go` nên đọc DSN từ env hoặc config file, không hardcode user/password trong code. Xem [internal/database/postgres.go](internal/database/postgres.go#L11).
- `internal/mapper/project_mapper.go::UpdateProjectRequestToEntity` nên bỏ nếu không dùng, hoặc làm nil-safe nếu muốn giữ lại. Hiện tại nó không đóng góp cho flow chính. Xem [internal/mapper/project_mapper.go](internal/mapper/project_mapper.go#L9).

## 5. Risk còn lại
- Security: log lộ password/hash, hardcode DB credentials, `/users` không có self-only protection.
- Logic: Register vẫn lệch thứ tự validate/duplicate check, Task update status casing chưa thống nhất, forbidden mapping của Task chưa chạy đúng.
- Database: register duplicate email vẫn phụ thuộc unique index; API cần map unique violation sang 409 rõ ràng.
- API permission: nếu requirement sau này chỉ cho user xem dữ liệu của chính họ thì project list-all và users CRUD hiện là rủi ro truy cập dữ liệu người khác.
- Documentation: README hiện có nguy cơ gây hiểu sai cách chạy project vì đang Outdated.

## 6. Checklist Pass / Partial / Fail

| Area | Status | Notes |
|---|---|---|
| Clean Architecture | Partial | Layering đúng hướng, nhưng error handling và logging còn lẫn trách nhiệm ở vài chỗ |
| Auth | Partial | Bcrypt/JWT có dùng, nhưng Register chưa chuẩn, log nhạy cảm còn tồn tại |
| User CRUD | Fail | Chưa có quyền self-only hoặc role-based protection; password log vẫn còn |
| Project CRUD | Partial | Update/delete có ownership check, nhưng thiếu CreateProject và get-by-id chưa map 404 |
| Task CRUD | Partial | Ownership check tốt, nhưng status casing và forbidden mapping còn lỗi |
| Database/Migration | Pass | Có PK/FK/UNIQUE/index/check constraint; cần bổ sung env-based config và 409 mapping |
| Error Handling | Partial | `ApiResponse` có chuẩn, nhưng `gin.H` raw và status code chưa đồng nhất |
| Unit Test | Fail | Test file chủ yếu comment out, chưa có test thực thi cho auth/task |
| README | Fail | Nội dung đang Outdated và thiếu hướng dẫn theo code hiện tại |

## 7. Priority Fixes
1. Xóa ngay toàn bộ log lộ password/hash trong AuthService và UserHandler.
2. Khóa lại `/users` endpoints theo requirement thật của dự án, tối thiểu là self-only nếu chỉ có role `USER`.
3. Sửa `AuthService.Register`: validate trước, dùng `ExistsByEmail`, và dừng đúng khi repository trả lỗi DB.
4. Tách status code của Register: 400 cho validation, 409 cho duplicate email.
5. Đồng bộ Task status giữa request validation và service validation, hoặc normalize status trước khi validate.
6. Sửa mapping not found/forbidden ở ProjectHandler và TaskHandler.
7. Đưa database config ra env và làm README khớp với code thực tế.
8. Viết unit test thật cho AuthService, TaskService, và các case duplicate/invalid/not found.

Nếu muốn, tôi có thể tiếp tục sửa trực tiếp các lỗi critical theo đúng thứ tự trên.