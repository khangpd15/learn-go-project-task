# Phân tích JWT Authentication/Authorization - Mentor Backend Go

Tài liệu này đọc theo code hiện tại trong repo và giải thích theo góc nhìn mentor cho người chuyển từ Java sang Golang.

## 0) Big Picture nhanh

Hệ thống của bạn đang đi đúng hướng kiến trúc nhiều layer:

- Route layer: định nghĩa endpoint và nhóm public/protected.
- Handler layer: HTTP I/O (bind request, trả response).
- Service layer: business logic (login, validate).
- Repository layer: truy vấn DB.
- Middleware layer: chặn request để xác thực/ủy quyền.

Luồng auth chuẩn bạn đang dùng:

1. Login với email/password.
2. So sánh password plaintext với password hash (bcrypt).
3. Sinh JWT token có claims (`user_id`, `email`, `exp`, `iat`).
4. Client gửi token qua `Authorization: Bearer <token>`.
5. Middleware verify token, nạp user từ DB, set context.
6. Request mới được đi vào protected handlers.

---

## 1) Giải thích cực chi tiết từng file auth/JWT liên quan

## 1.1 `internal/services/auth_service.go`

### Mục tiêu file

- Chứa business logic đăng nhập.
- Không xử lý HTTP trực tiếp.
- Không truy cập DB trực tiếp, mà đi qua repository interface.

### Giải thích theo từng dòng/khối

- `type AuthServiceInterface interface { Login(req auth.LoginRequest) (string, error) }`
  - Định nghĩa contract cho service auth.
  - Lợi ích: dễ mock khi test, giảm coupling.

- `type AuthService struct { userRepo repositories.UserRepositoryInterface }`
  - Service phụ thuộc abstraction (`UserRepositoryInterface`) thay vì struct cụ thể.
  - Đây là dependency inversion (rất giống tư duy Spring dùng interface).

- `NewAuthService(...)`
  - Constructor inject dependency.
  - Tránh tạo repository bên trong service, giúp test và maintain dễ.

- `func (as *AuthService) Login(req auth.LoginRequest) (string, error)`
  - Entry point nghiệp vụ login.

- `user, err := as.userRepo.GetUserByEmail(req.Email)`
  - Tìm user theo email.
  - Nếu không thấy user: trả generic error để chống email enumeration (đúng hướng).

- `if err != nil { return "", errors.New("invalid email or password") }`
  - Không lộ "email not found" hay "password sai" riêng lẻ.
  - Đây là best practice bảo mật.

- `fmt.Println(...)` các dòng debug email/hash/password
  - Dùng để debug tạm thời.
  - Trong production là rủi ro nghiêm trọng vì lộ dữ liệu nhạy cảm (đặc biệt password plaintext).

- `err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))`
  - So sánh hash lưu DB với password người dùng gửi lên.
  - Hàm bcrypt tự extract salt/cost từ hash rồi hash lại plaintext để compare.

- `if err != nil { return "", errors.New("invalid email or password") }`
  - Password sai -> login fail với message generic.

- `token, err := utils.GenerateAccessToken(user.ID, user.Email)`
  - Tạo JWT access token từ user info.

- `if err != nil { return "", errors.New("failed to generate access token") }`
  - Bọc lỗi để handler trả response ổn định.

- `return token, nil`
  - Thành công: trả token cho handler.

### Vì sao viết như vậy

- Service chịu trách nhiệm nghiệp vụ, handler chỉ điều phối HTTP.
- So sánh password hash thay vì plaintext là bắt buộc về security.
- Dùng repository interface để loose coupling.

---

## 1.2 `internal/handler/auth_handler.go`

### Mục tiêu file

- Nhận request HTTP login.
- Bind JSON input.
- Gọi service.
- Trả JSON response chuẩn.

### Giải thích theo từng dòng/khối

- `type AuthHandler struct { authService *services.AuthService }`
  - Handler giữ service để gọi nghiệp vụ.

- `NewAuthHandler(...)`
  - Constructor inject service.

- `func (h *AuthHandler) Login(c *gin.Context)`
  - Gin handler cho endpoint login.

- `var req requestDTO.LoginRequest`
  - Khai báo DTO request.

- `c.ShouldBindJSON(&req)`
  - Parse body JSON vào struct.
  - Nếu fail -> `400 Bad Request`.

- `token, err := h.authService.Login(req)`
  - Delegation sang service.

- `if err != nil { c.JSON(401, ...) }`
  - Login fail trả 401.

- `c.JSON(200, response.SuccessResponse(..., gin.H{"access_token": token}))`
  - Login success trả access token cho client.

- `LogOut(...)`
  - Hiện tại chỉ trả success giả lập.
  - JWT stateless nên logout thật thường cần blacklist token hoặc rotate signing key/refresh token strategy.

### Vì sao viết như vậy

- Handler không chứa business logic phức tạp.
- Response format nhất quán qua `response.SuccessResponse/ErrorResponse`.

---

## 1.3 `internal/routes/auth_routes.go`

### Mục tiêu file

- Khai báo endpoint auth.

### Giải thích theo từng dòng/khối

- `type AuthRoutes struct { authHandler *handler.AuthHandler }`
  - Route struct giữ handler để map endpoint -> function.

- `SetupAuthRoutes(router *gin.RouterGroup)`
  - Tạo subgroup `/auth`.

- `authGroup.POST("/login", r.authHandler.Login)`
  - Public endpoint login.

### Vì sao viết như vậy

- Tách route theo module (`auth`, `task`, `user`) giúp dễ scale.

---

## 1.4 `internal/middleware/auth.go` (auth_middleware)

Bạn gọi tên là `auth_middleware.go`; trong repo hiện tại file tương đương là `auth.go`.

### Mục tiêu file

- Kiểm tra token ở mọi protected routes.
- Trích user từ token + DB.
- Gắn user vào request context.

### Giải thích theo từng dòng/khối

- `func AuthMiddleware(userRepo repositories.UserRepositoryInterface) gin.HandlerFunc`
  - Middleware nhận dependency userRepo (để query user theo `user_id` trong token).

- `authHeader := c.GetHeader("Authorization")`
  - Đọc header Authorization.

- `if authHeader == "" { ... c.Abort(); return }`
  - Thiếu header -> 401, dừng pipeline bằng `c.Abort()`.

- `tokenString := strings.TrimPrefix(authHeader, "Bearer ")`
  - Cắt tiền tố `Bearer `.

- `if tokenString == authHeader { ... }`
  - Nếu trim không thay đổi nghĩa là format sai (không có prefix chuẩn).

- `claims, err := utils.ValidateAccessToken(tokenString)`
  - Verify chữ ký token + parse claims.

- `userIDFloat, ok := claims["user_id"].(float64)`
  - Lấy `user_id` từ claims map.
  - Vì JSON number mặc định decode thành `float64`, nên phải assert kiểu này trước rồi cast `int`.

- `userID := int(userIDFloat)`
  - Convert sang int để query DB.

- `user, err := userRepo.GetUserByID(userID)`
  - Nạp user từ DB để đảm bảo user còn tồn tại.

- `c.Set("current_user", user)`
  - Lưu object user vào gin context cho downstream handlers/services.

- `c.Set("user_id", user.ID)`
  - Lưu user_id riêng để lấy nhanh.

- `c.Next()`
  - Cho request đi tiếp tới handler kế tiếp.

### Vì sao viết như vậy

- Middleware giúp tách logic auth khỏi từng handler để tránh lặp code.
- Query DB sau khi verify token giúp "token hợp lệ nhưng user bị xoá/disable" không qua được.

---

## 1.5 `internal/utils/jwt.go`

### Mục tiêu file

- Sinh JWT và validate JWT.

### Giải thích theo từng dòng/khối

- `var jwtSecretKey = []byte("your_secret_key")`
  - Key ký token HMAC.
  - Hiện tại hardcode -> cần chuyển env var trong production.

- `GenerateAccessToken(userID int, email string)`
  - Tạo claims map gồm:
  - `user_id`: id người dùng.
  - `email`: email người dùng.
  - `exp`: thời điểm hết hạn (Unix timestamp).
  - `iat`: thời điểm phát hành token.

- `token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)`
  - Tạo JWT dùng thuật toán ký HS256.

- `return token.SignedString(jwtSecretKey)`
  - Ký token và serialize thành chuỗi `header.payload.signature`.

- `ValidateAccessToken(tokenString string)`
  - Parse token string.

- callback `func(token *jwt.Token) (interface{}, error)`
  - Validate thuật toán ký mong muốn (HS256).
  - Trả về secret key để thư viện verify chữ ký.

- `claims, ok := token.Claims.(jwt.MapClaims)` + `token.Valid`
  - Kiểm tra claims parse được và token hợp lệ.

- `return claims, nil`
  - Trả claims để middleware dùng.

### Vì sao viết như vậy

- Gom logic JWT vào utility dùng chung.
- Tách khỏi handler/service giúp tái sử dụng và test độc lập.

---

## 1.6 `internal/repositories/user_repository.go` (GetUserByEmail / GetUserByID)

### GetUserByEmail

- Query `WHERE email = ?`.
- Nếu `gorm.ErrRecordNotFound` thì map thành lỗi `user not found`.
- Trả `*entities.User`.

Tại sao quan trọng với auth:

- Đây là điểm vào đầu tiên của login.
- Nếu thiếu index email sẽ chậm, nhưng migration đã có unique/index phù hợp.

### GetUserByID

- Query `WHERE id = ?`.
- Middleware dùng hàm này sau khi đọc `user_id` từ JWT.

Tại sao cần query DB lại thay vì chỉ tin JWT:

- Token có thể còn hạn nhưng user đã bị xoá/khóa.
- DB check là lớp bảo vệ runtime quan trọng.

---

## 1.7 `internal/entities/user.go`

### Mục tiêu file

- Mô hình user dùng xuyên suốt app.

### Giải thích theo từng field

- `ID int`: khoá chính.
- `FullName string` với tag `json:"fullname"`: tên đầy đủ.
- `PasswordHash string` với `json:"-"`: không bao giờ trả ra JSON response.
- `Email string`: email login.
- `CreatedAt time.Time`: timestamp tạo user.

### Lưu ý quan trọng

- Migration dùng cột `full_name`; entity chỉ có json tag, chưa có gorm tag cho `full_name`.
- `PasswordHash` có gorm tag đúng (`column:password_hash`).
- Nếu naming strategy mặc định không map đúng `FullName -> full_name`, có thể sinh bug data mapping.

---

## 1.8 `internal/routes/routes.go` (public/protected routes)

### Flow setup

- Tạo `v1 := r.Group("/api/v1")`.
- Public: `NewAuthRoutes(authHandler).SetupAuthRoutes(v1)`.
- Protected:
  - `protected := v1.Group("")`
  - `protected.Use(middleware.AuthMiddleware(userRepo))`
  - Task routes + User routes gắn vào protected.

Ý nghĩa:

- `/api/v1/auth/login` không cần token.
- Các route `/api/v1/tasks/*`, `/api/v1/users/*` bắt buộc có Bearer token.

---

## 1.9 Create task flow dùng JWT current user

## Trạng thái hiện tại trong code

Hiện tại `TaskHandler.CreateTask` bind trực tiếp JSON vào `entities.Task`:

- Không đọc `user_id` từ context.
- Client có thể tự gửi `assignee_id` bất kỳ.
- Chưa có logic "auto assign current logged-in user".

Nói ngắn gọn: middleware đã set `current_user`/`user_id`, nhưng create task chưa dùng.

## Nên làm đúng như JWT flow

Trong handler create task:

1. Lấy `user_id` từ `c.Get("user_id")`.
2. Ép kiểu sang int.
3. Gán vào `task.AssigneeID` trước khi gọi service/repo.
4. Không tin `assignee_id` client gửi lên (ít nhất trong luồng user tự tạo task cho mình).

---

## 1.10 Middleware context `current_user` và `user_id`

### `gin.Context` là gì

- Object context cho mỗi HTTP request trong Gin.
- Chứa request/response metadata + nơi truyền dữ liệu giữa middleware và handler.

### `c.Set()` và `c.Get()`

- `c.Set(key, value)`: ghi dữ liệu vào context map của request hiện tại.
- `c.Get(key)`: đọc lại dữ liệu đã set ở middleware trước đó.

Vì sao cần:

- Tránh parse token lặp lại trong từng handler.
- Truyền current user xuyên suốt pipeline request.

---

## 2) Giải thích các khái niệm bắt buộc

## 2.1 JWT là gì

JWT (JSON Web Token) là chuỗi token gồm 3 phần:

- Header: thuật toán ký, loại token.
- Payload: claims (dữ liệu như user_id, email, exp...).
- Signature: chữ ký số để chống sửa payload.

JWT thường dùng cho stateless auth trong REST API.

## 2.2 Claims là gì

Claims là dữ liệu trong payload JWT.

- Registered claims: `exp`, `iat`, `nbf`, `iss`, `aud`, `sub`...
- Private claims: `user_id`, `email`, `role`...

## 2.3 `exp` và `iat`

- `exp` (expiration time): thời điểm token hết hạn. Sau mốc này token phải bị từ chối.
- `iat` (issued at): thời điểm token được phát hành. Hữu ích cho audit, invalidate theo mốc, debug.

## 2.4 bcrypt hoạt động ra sao

bcrypt:

1. Tạo salt ngẫu nhiên.
2. Hash password + salt nhiều vòng (cost).
3. Lưu chuỗi hash đã chứa thông tin salt + cost.

Khi login:

- `CompareHashAndPassword` lấy cost/salt từ hash cũ.
- Hash lại plaintext input.
- So sánh kết quả.

## 2.5 Tại sao không lưu plaintext password

- Nếu DB lộ, plaintext = lộ toàn bộ tài khoản ngay lập tức.
- Hash một chiều giúp giảm thiệt hại.
- bcrypt chống brute force tốt hơn hash nhanh như SHA-256 thuần.

## 2.6 Tại sao dùng middleware

- Tái sử dụng logic auth cho nhiều route.
- Single source of truth cho auth check.
- Giữ handler sạch và tập trung business.

## 2.7 Tại sao dùng Authorization Bearer Token

- Chuẩn RFC6750 cho OAuth2/Bearer.
- Tương thích hầu hết reverse proxy/API gateway/client tools.
- Dễ tích hợp frontend/mobile/Postman.

## 2.8 Tại sao không dùng User-ID header nữa

- `User-ID: 5` có thể bị giả mạo cực dễ.
- Không có chữ ký, không có expiry.
- Không có non-repudiation.

Bearer JWT có chữ ký số và thời hạn, an toàn hơn nhiều.

---

## 3) Giải thích middleware theo đúng 8 bước bạn yêu cầu

1. Đọc Authorization header
- `c.GetHeader("Authorization")`.

2. Validate format Bearer token
- Dùng `strings.TrimPrefix(authHeader, "Bearer ")`.
- Nếu không có prefix -> reject.

3. Verify JWT
- `utils.ValidateAccessToken(tokenString)` parse + verify chữ ký + thuật toán.

4. Lấy claims
- Token hợp lệ -> nhận `jwt.MapClaims`.

5. Lấy user_id
- `claims["user_id"].(float64)` rồi cast `int`.

6. Query user từ DB
- `userRepo.GetUserByID(userID)`.

7. Set current_user vào context
- `c.Set("current_user", user)` và `c.Set("user_id", user.ID)`.

8. Cho request đi tiếp bằng `c.Next()`
- Handler business bắt đầu chạy.

---

## 4) Vì sao `user_id` trong JWT ra `float64`

Vì `jwt.MapClaims` bản chất là `map[string]interface{}` và JSON number mặc định unmarshal thành `float64` trong Go.

Nên pattern thường thấy:

- assert `float64` trước.
- validate value nguyên dương.
- cast về `int` để dùng nội bộ.

Nếu muốn type-safe hơn:

- Dùng custom claims struct thay vì `MapClaims`.

---

## 5) Giải thích các hàm cụ thể

## 5.1 `strings.TrimPrefix()`

- Cắt prefix nếu có.
- Nếu không có, trả nguyên chuỗi cũ.
- Trong middleware dùng để tách token khỏi chuỗi `Bearer <token>`.

## 5.2 `bcrypt.CompareHashAndPassword()`

- Input 1: hash đã lưu.
- Input 2: password plaintext user nhập.
- Trả `nil` nếu khớp, lỗi nếu không khớp.

## 5.3 `token.SignedString()`

- Dùng secret/private key để ký token.
- Kết quả là JWT string gửi cho client.
- Nếu ký thất bại phải trả lỗi ngay.

---

## 6) Luồng request end-to-end (Login -> Bearer -> Protected)

## 6.1 Login flow

Client -> AuthHandler.Login -> AuthService.Login -> UserRepository.GetUserByEmail -> DB

Nếu pass:

AuthService -> utils.GenerateAccessToken -> trả token -> AuthHandler -> JSON response `access_token`

## 6.2 Protected API flow

Client gửi `Authorization: Bearer <token>` -> AuthMiddleware

AuthMiddleware:

- Parse header.
- Verify token.
- Lấy user_id claim.
- Query DB user.
- Set context.
- `c.Next()`.

Sau đó mới vào Task/User Handler.

---

## 7) Flow text diagram

## 7.1 Business flow tổng quát

Client
-> Route
-> Handler
-> Service
-> Repository
-> DB
-> Repository
-> Service
-> Handler
-> Client

## 7.2 JWT middleware flow

Client request (Bearer token)
-> Gin Router
-> AuthMiddleware
-> Read Authorization header
-> Validate Bearer format
-> Validate JWT signature + exp
-> Extract claims user_id
-> Query user by ID
-> Set current_user/user_id vào context
-> c.Next()
-> Protected Handler

---

## 8) Đánh giá production-readiness: phần ổn, refactor, security risk

## 8.1 Code đang ổn

- Kiến trúc layer tách tương đối rõ.
- Có dùng bcrypt hash/compare.
- Có JWT exp/iat.
- Có route group public/protected.
- Middleware có DB check user tồn tại.
- Password hash không lộ ra JSON response (`json:"-"`).

## 8.2 Nên refactor sớm

1. Xóa log nhạy cảm trong auth/user handlers
- `fmt.Println` đang in plaintext password/hash.

2. Dùng interface ở handler/service nhất quán
- `AuthHandler` đang phụ thuộc concrete `*services.AuthService`, có thể chuyển sang interface cho dễ test.

3. Dùng typed claims thay vì MapClaims
- Tránh cast `float64` thủ công.

4. Chuẩn hóa DTO create task
- Không bind trực tiếp `entities.Task` từ request.
- Dùng request DTO rõ field cho phép client nhập.

5. Tách logout strategy rõ ràng
- Nếu stateless thuần: logout client-side.
- Nếu cần revoke: thêm token blacklist store hoặc refresh-token rotation.

6. Chuẩn hóa naming file
- Có thể rename `internal/middleware/auth.go` -> `auth_middleware.go` cho đồng bộ với convention team.

## 8.3 Vấn đề bảo mật tiềm ẩn

1. JWT secret hardcode
- `your_secret_key` trong source code là rủi ro cao.

2. DSN DB hardcode username/password
- Lộ credential trong codebase.

3. Không có prefix check case-insensitive/space normalize cho Bearer
- Có thể siết parsing chặt hơn.

4. Chưa có refresh token
- Access token sống 24h có thể hơi dài tùy threat model.

5. Chưa có role/authorization thật
- `role.go` đang comment out.
- Hiện tại xác thực có user, nhưng chưa kiểm soát quyền tài nguyên.

6. Create task chưa cưỡng ép assignee theo current user
- Có thể bị spoof assignee.

## 8.4 Best practices tiếp theo

- Secret/config qua env + secret manager.
- Access token ngắn (15-30 phút) + refresh token.
- Custom claims struct + validate chuẩn (`exp`, `iat`, `nbf`, `iss`, `aud`).
- Đưa auth errors về schema chuẩn không lộ thông tin.
- Implement RBAC (role claims + role middleware).
- Audit logging an toàn (không log password/token raw).
- Viết integration test cho login + protected routes + expired token.

---

## 9) Ví dụ Postman request/response

## 9.1 Login

Request:

- Method: POST
- URL: `/api/v1/auth/login`
- Headers:
  - `Content-Type: application/json`
- Body:

```json
{
  "email": "admin@example.com",
  "password": "StrongP@ss1"
}
```

Response thành công (200):

```json
{
  "status": true,
  "message": "Login successfully",
  "data": {
    "access_token": "<jwt_token_here>"
  }
}
```

Response thất bại (401):

```json
{
  "status": false,
  "message": "Login failed",
  "error": "invalid email or password"
}
```

## 9.2 Get tasks (protected)

Request:

- Method: GET
- URL: `/api/v1/tasks`
- Headers:
  - `Authorization: Bearer <access_token>`

Response (200):

```json
{
  "status": true,
  "message": "get all tasks successfully",
  "data": [
    {
      "id": 1,
      "project_id": 1,
      "title": "Task A",
      "description": "...",
      "status": "TODO",
      "assignee_id": 2,
      "created_at": "2026-05-19T10:00:00Z"
    }
  ]
}
```

Nếu thiếu token (401):

```json
{
  "message": "missing Authorization header"
}
```

## 9.3 Create task (protected)

Request hiện tại của code:

- Method: POST
- URL: `/api/v1/tasks`
- Headers:
  - `Authorization: Bearer <access_token>`
  - `Content-Type: application/json`
- Body:

```json
{
  "project_id": 1,
  "title": "Implement JWT",
  "description": "Add middleware and claims",
  "status": "TODO",
  "assignee_id": 2
}
```

Response (201):

```json
{
  "status": true,
  "message": "Task created successfully",
  "data": {
    "id": 10,
    "project_id": 1,
    "title": "Implement JWT",
    "description": "Add middleware and claims",
    "status": "TODO",
    "assignee_id": 2,
    "created_at": "2026-05-19T10:10:00Z"
  }
}
```

Khuyến nghị production:

- Không cho client gửi `assignee_id` trong luồng "tạo task của chính mình".
- Server tự set `assignee_id = user_id` từ JWT context.

---

## 10) Giải thích kiến trúc cho người từ Java sang Go

## 10.1 Repository pattern

- Java (Spring Data): `UserRepository extends JpaRepository`.
- Go của bạn: interface + struct GORM thực thi thủ công.
- Ưu điểm: explicit, dễ hiểu luồng query, dễ mock.

## 10.2 Service layer

- Java: `@Service`.
- Go: struct service + constructor.
- Chứa business rule, không phụ thuộc HTTP framework.

## 10.3 Handler layer

- Java: `@RestController`.
- Go: Gin handler method nhận `*gin.Context`.
- Chỉ nên làm việc HTTP (bind, status code, response).

## 10.4 Middleware layer

- Java: filter/interceptor.
- Go Gin: middleware chain (`Use(...)`).
- Chạy trước handler; có thể `Abort` để chặn request.

## 10.5 Route layer

- Java: annotation mapping.
- Go Gin: map bằng code (`GET`, `POST`, group).
- Tách module route rõ ràng giúp maintain tốt.

---

## 11) Checklist cải tiến ưu tiên (theo thứ tự)

1. Xóa toàn bộ log plaintext password/hash ngay.
2. Đưa JWT secret + DB DSN sang env.
3. Sửa create task để dùng `user_id` từ context.
4. Tạo DTO riêng cho create task, không bind entity trực tiếp.
5. Bổ sung role-based authorization thật (RBAC).
6. Thêm refresh token strategy và revoke flow.
7. Viết test cho auth middleware + token hết hạn + token sai chữ ký.

---

## 12) Kết luận mentor

Bạn đã có nền tảng auth flow đúng hướng production ở mức cơ bản: login bằng bcrypt, phát JWT, chặn protected routes bằng middleware và nạp current user vào context.

Khoảng cách lớn nhất để lên production-ready là:

- Secret/config management.
- Loại bỏ logging nhạy cảm.
- Ràng buộc ownership/authorization mạnh hơn (đặc biệt create/update/delete task).
- Chuẩn hóa token lifecycle (refresh/revoke) và claims typing.

Nếu bạn muốn, bước tiếp theo mình có thể tạo ngay patch code để:

- Auto-assign `assignee_id` = current user khi create task.
- Chặn client override assignee_id.
- Bỏ log nhạy cảm.
- Đưa JWT secret sang biến môi trường.
