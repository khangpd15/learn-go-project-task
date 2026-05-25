# Postman Test Plan - Task Management API

Tài liệu này dùng để test toàn bộ chức năng API bằng Postman theo đúng behavior hiện tại của codebase.

## Kết luận nhanh về trạng thái code

Code hiện tại đủ để bắt đầu functional test bằng Postman.

Những điểm đã ổn hơn trước:
- Auth middleware đã trả response theo format chuẩn thay vì raw map.
- DTO, mapper và response layer đã tách rõ hơn.
- Task status update flow đã khớp giữa request validation và service.
- User duplicate email đã map được sang HTTP 409 ở handler.

Những điểm vẫn cần lưu ý khi test:
- `POST /api/v1/auth/register` vẫn trả HTTP 400 cho duplicate email theo service/handler hiện tại.
- `GET /api/v1/projects/:id` vẫn map lỗi repository chung chung sang HTTP 500.
- `GET /api/v1/tasks/:id` và một vài flow task/project vẫn phụ thuộc vào error string thay vì sentinel error đầy đủ.
- `RecoveryMiddleware` vẫn trả raw `gin.H` nếu panic xảy ra.

## Môi trường test

- Base URL: `http://localhost:8080`
- Postgres và Redis đã chạy qua Docker Compose.
- Cần có `JWT_SECRET` trong môi trường chạy app.

## Postman setup

### Environment variables

Tạo environment trong Postman với các biến sau:

- `baseUrl` = `http://localhost:8080`
- `accessToken` = rỗng lúc đầu
- `userId` = rỗng
- `projectId` = rỗng
- `taskId` = rỗng

### Headers dùng chung

- `Content-Type: application/json`
- `Authorization: Bearer {{accessToken}}` cho toàn bộ endpoint protected

## Test flow đề xuất

1. Health check
2. Register user
3. Login và lưu token
4. Test các endpoint users
5. Test các endpoint projects
6. Test các endpoint tasks
7. Test negative cases: invalid token, invalid body, forbidden, not found

## Test cases chi tiết

### 1. Health Check

#### 1.1 GET /health

- Method: `GET`
- URL: `{{baseUrl}}/health`
- Auth: không cần

Expected:
- HTTP 200
- Body có `message: server is running`

---

### 2. Auth

#### 2.1 Register user thành công

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/register`
- Auth: không cần

Body:
```json
{
  "fullname": "Test User",
  "email": "test.user@example.com",
  "password": "Password@123"
}
```

Expected:
- HTTP 200
- Body `status = true`
- Message: `Register successfully`

#### 2.2 Register với email trùng

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/register`

Body: dùng lại email ở trên

Expected theo code hiện tại:
- HTTP 400
- Body `status = false`
- Message: `Register failed`
- Error: `email already exists`

Ghi chú:
- Về best practice nên là HTTP 409, nhưng hiện tại code đang trả 400.

#### 2.3 Login thành công

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/login`
- Auth: không cần

Body:
```json
{
  "email": "test.user@example.com",
  "password": "Password@123"
}
```

Expected:
- HTTP 200
- Body `status = true`
- `data.access_token` có giá trị

Postman test script gợi ý:
```javascript
const json = pm.response.json();
pm.environment.set("accessToken", json.data.access_token);
```

#### 2.4 Login sai password

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/login`

Body:
```json
{
  "email": "test.user@example.com",
  "password": "WrongPassword@123"
}
```

Expected:
- HTTP 401
- Message: `Login failed`
- Error: `invalid email or password`

#### 2.5 Login email không tồn tại

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/login`

Body:
```json
{
  "email": "missing@example.com",
  "password": "Password@123"
}
```

Expected:
- HTTP 401
- Message: `Login failed`
- Error: `invalid email or password`

---

### 3. Users

#### 3.1 GET /users - lấy danh sách users

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/users`
- Auth: có token

Expected:
- HTTP 200
- Body `status = true`
- `data` là array user response

#### 3.2 GET /users/:id - lấy user theo ID

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/users/{{userId}}`
- Auth: có token

Expected:
- HTTP 200 nếu tồn tại
- Body trả `id`, `full_name`, `email`

Negative case:
- ID không hợp lệ: HTTP 400
- ID không tồn tại: HTTP 404

#### 3.3 GET /users/email/:email

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/users/email/test.user@example.com`
- Auth: có token

Expected:
- HTTP 200 nếu tồn tại
- HTTP 404 nếu không tồn tại

Lưu ý:
- Code hiện tại có thể trả sai status nếu repository không chuẩn hóa error not-found.

#### 3.4 GET /users/fullname/:fullname

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/users/fullname/Test%20User`
- Auth: có token

Expected:
- HTTP 200 nếu tồn tại
- HTTP 404 nếu không tồn tại

#### 3.5 POST /users - tạo user mới

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/users`
- Auth: có token

Body:
```json
{
  "full_name": "New User",
  "email": "new.user@example.com",
  "password": "Password@123"
}
```

Expected:
- HTTP 201
- Body `status = true`
- Không trả plaintext password

#### 3.6 PUT /users/:id - cập nhật user

- Method: `PUT`
- URL: `{{baseUrl}}/api/v1/users/{{userId}}`
- Auth: có token

Body:
```json
{
  "full_name": "Updated User",
  "email": "updated.user@example.com",
  "password": "Password@123"
}
```

Expected:
- HTTP 200
- `full_name` và `email` cập nhật đúng
- Password mới được hash trong DB

#### 3.7 DELETE /users/:id

- Method: `DELETE`
- URL: `{{baseUrl}}/api/v1/users/{{userId}}`
- Auth: có token

Expected:
- HTTP 200
- User bị xóa

---

### 4. Projects

#### 4.1 GET /projects/me

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/projects/me`
- Auth: có token

Expected:
- HTTP 200
- Chỉ trả project thuộc current user

#### 4.2 GET /projects - list all projects

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/projects`
- Auth: có token

Expected theo code hiện tại:
- HTTP 200
- Trả toàn bộ projects

Lưu ý:
- Nếu spec muốn giới hạn theo owner, endpoint này đang quá mở.

#### 4.3 GET /projects/:id

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/projects/{{projectId}}`
- Auth: có token

Expected:
- HTTP 200 nếu tồn tại
- HTTP 404 nếu không tồn tại

Lưu ý:
- Code hiện tại vẫn có thể trả 500 trong một số trường hợp not-found/DB error lẫn nhau.

#### 4.4 PUT /projects/:id

- Method: `PUT`
- URL: `{{baseUrl}}/api/v1/projects/{{projectId}}`
- Auth: có token

Body:
```json
{
  "name": "Updated Project",
  "description": "Updated description"
}
```

Expected:
- HTTP 200 nếu owner đúng
- HTTP 403 nếu project không thuộc user hiện tại
- HTTP 404 nếu project không tồn tại

#### 4.5 DELETE /projects/:id

- Method: `DELETE`
- URL: `{{baseUrl}}/api/v1/projects/{{projectId}}`
- Auth: có token

Expected:
- HTTP 200 nếu owner đúng
- HTTP 403 nếu project không thuộc user hiện tại
- HTTP 404 nếu project không tồn tại

---

### 5. Tasks

#### 5.1 GET /tasks

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/tasks`
- Auth: có token

Expected:
- HTTP 200
- Chỉ trả tasks thuộc projects của user hiện tại

#### 5.2 GET /tasks/:id

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/tasks/{{taskId}}`
- Auth: có token

Expected:
- HTTP 200 nếu task thuộc project của user hiện tại
- HTTP 403 nếu task thuộc user khác
- HTTP 404 nếu task không tồn tại

#### 5.3 POST /tasks

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/tasks`
- Auth: có token

Body:
```json
{
  "project_id": 1,
  "title": "First task",
  "description": "Task description"
}
```

Expected:
- HTTP 201
- `status = true`
- Task được tạo với status mặc định `TODO`
- `assignee_id` được set theo current user

Negative cases:
- Project ID không thuộc user hiện tại: HTTP 403
- Project không tồn tại: HTTP 404
- Body thiếu field required: HTTP 400

#### 5.4 PUT /tasks/:id

- Method: `PUT`
- URL: `{{baseUrl}}/api/v1/tasks/{{taskId}}`
- Auth: có token

Body:
```json
{
  "title": "Updated task",
  "description": "Updated desc",
  "status": "done"
}
```

Expected:
- HTTP 200 nếu owner đúng
- Status được normalize thành uppercase trong service
- Trong DB lưu `DONE`

Negative cases:
- `status` không hợp lệ: HTTP 400
- task không thuộc user hiện tại: HTTP 403
- task không tồn tại: HTTP 404

#### 5.5 DELETE /tasks/:id

- Method: `DELETE`
- URL: `{{baseUrl}}/api/v1/tasks/{{taskId}}`
- Auth: có token

Expected:
- HTTP 200 nếu owner đúng
- HTTP 403 nếu task không thuộc user hiện tại
- HTTP 404 nếu task không tồn tại

---

### 6. Negative / Security tests

#### 6.1 Missing Authorization header

- Gọi bất kỳ protected endpoint nào không set token

Expected:
- HTTP 401
- Error message về authorization header

#### 6.2 Bearer token sai format

- Header: `Authorization: Token abc`

Expected:
- HTTP 401
- Error message: bearer format invalid

#### 6.3 Token sai hoặc hết hạn

- Dùng token random hoặc token hết hạn

Expected:
- HTTP 401

#### 6.4 Body invalid JSON

- Gửi JSON lỗi cú pháp hoặc thiếu field required

Expected:
- HTTP 400

#### 6.5 ID không phải số

- Ví dụ `/users/abc`, `/tasks/abc`, `/projects/abc`

Expected:
- HTTP 400

---

## Checklist expected results

- Register works with valid body.
- Login returns access token.
- Protected routes reject missing/invalid token.
- User CRUD returns sanitized response without password.
- Project ownership rules are enforced on update/delete.
- Task ownership rules are enforced on read/update/delete.
- Task status update accepts lowercase input and stores uppercase.
- Duplicate email is handled consistently enough to test, but the current status code is still not ideal.

## Notes for reviewer / tester

- Nếu endpoint nào trả HTTP 500 thay vì 404/403, đó thường là dấu hiệu error mapping chưa chuẩn hóa, không phải logic DB hỏng.
- Nếu Postman test script cần lưu `projectId` hoặc `taskId`, bạn có thể lấy từ response `data.id` của project/task create.
- Nên test theo thứ tự: auth → users → projects → tasks, vì tasks và projects phụ thuộc ownership/token.
