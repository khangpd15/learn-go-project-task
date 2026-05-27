# Phân Tích Kiến Trúc Hiện Tại Và Khả Năng Tích Hợp Realtime

Tài liệu này dựa trên source hiện tại trong [cmd/api/main.go](cmd/api/main.go), [internal/app/app.go](internal/app/app.go), [internal/routes](internal/routes), [internal/handler](internal/handler), [internal/services](internal/services), [internal/repositories](internal/repositories), [internal/middleware](internal/middleware), [internal/entities](internal/entities), [internal/dto](internal/dto), [internal/response](internal/response), [internal/config](internal/config), [internal/utils](internal/utils), [internal/cache](internal/cache), [internal/queue](internal/queue), [internal/jobs](internal/jobs), [internal/worker](internal/worker), và các migration / test hiện có.

Mục tiêu ở đây là phân tích kiến trúc hiện tại, chứ chưa triển khai WebSocket.

## 1. Tổng Quan Kiến Trúc Hiện Tại

Project đang đi theo layered monolith, với một composition root rõ tại [internal/app/app.go](internal/app/app.go). Có thể nhìn thấy các lớp chính:

- Transport layer: routes, middleware, handler.
- Application layer: services.
- Persistence layer: repositories, database.
- Shared model layer: entities, DTO, mapper, validation.
- Infrastructure layer: cache, queue, worker, config.

Clean Architecture hiện mới ở mức bán phần. Điểm đúng là handler không chứa SQL, service không phụ thuộc Gin, repository không biết về HTTP, và DTO đã được tách khỏi entity. Tuy nhiên, mức phân ranh giới vẫn chưa đủ chặt để gọi là Clean Architecture đúng nghĩa, vì business semantics vẫn bị rò vào repository và một số cross-cutting concerns vẫn xử lý trực tiếp trong service.

Dependency flow hiện tại là:

HTTP request
→ middleware
→ handler
→ service
→ repository
→ database / cache / queue

Điểm tốt:

- Có composition root tập trung ở [internal/app/app.go](internal/app/app.go).
- JWT auth được tách ra middleware riêng.
- DTO và entity đã tách bạch.
- Cache và queue đã được abstraction hóa bằng interface, nên có đường cho realtime / worker sau này.
- Service layer đã bắt đầu enforce ownership cho project/task.

Điểm chưa production-ready:

- Auth middleware đang vừa xác thực token vừa query DB để lấy user mỗi request, nên chi phí auth cao hơn cần thiết.
- Recovery middleware hiện giả định panic là string, có thể làm phát sinh panic thứ cấp nếu panic là kiểu khác.
- JWT secret được đọc ở package init time trong [internal/utils/jwt.go](internal/utils/jwt.go), nên cấu hình runtime chưa thật chắc.
- Route `/users` vẫn mở CRUD cho mọi authenticated user, chưa có self-only hoặc role check.
- Một số debug output và logging trong service còn mang tính phát triển, chưa phù hợp production.
- Repository đang chứa một phần logic nghiệp vụ theo ngữ nghĩa ownership / assignment, chưa thuần persistence.
- Response handling chưa đồng nhất: có chỗ trả mapper, có chỗ trả entity thô.

## 2. Flow Request Hiện Tại

### 2.1 Entry point

Flow bắt đầu từ [cmd/api/main.go](cmd/api/main.go), file này chỉ gọi [internal/app.Run](internal/app/app.go). Ở đây app khởi tạo Postgres, Redis, cache, queue, repository, service, handler, rồi gắn routes vào Gin engine.

### 2.2 Auth flow

Ví dụ login:

```text
POST /api/v1/auth/login
→ AuthHandler.Login
→ AuthService.Login
→ UserRepository.GetUserByEmail
→ bcrypt.CompareHashAndPassword
→ utils.GenerateAccessToken
→ JSON response
```

Register cũng đi tương tự, nhưng thêm validate email/password, check tồn tại email, hash password, rồi create user.

### 2.3 Protected flow

Protected routes được nhóm dưới `v1.Group("")` trong [internal/routes/routes.go](internal/routes/routes.go) và gắn [middleware.AuthMiddleware](internal/middleware/auth.go).

Mẫu flow điển hình:

```text
Client
→ AuthMiddleware
→ Handler
→ Service
→ Repository
→ Database
→ Response mapper
```

AuthMiddleware làm các việc sau:

- Đọc header Authorization.
- Parse Bearer token.
- Gọi `utils.ValidateAccessToken`.
- Lấy `user_id` từ claims.
- Query user thật từ DB qua `UserRepository.GetUserByID`.
- Set `current_user` và `user_id` vào Gin context.

Điều này có nghĩa là mọi request protected đều phải pass qua token validation và DB lookup trước khi tới handler.

### 2.4 Task flow thực tế

Ví dụ `GET /api/v1/tasks`:

```text
AuthMiddleware
→ TaskHandler.GetAllTasks
→ utils.CurrentUserID(c)
→ TaskService.GetAllTasks(currentUserID, ctx)
→ cache lookup by user key
→ cache miss thì TaskRepository.GetAllTasksByUserID
→ cache save
→ mapper.TasksToResponses
→ response.SuccessResponse
```

Ví dụ `POST /api/v1/tasks`:

```text
AuthMiddleware
→ TaskHandler.CreateTask
→ bind JSON
→ mapper.CreateTaskRequestToTaskEntity
→ TaskService.CreateTask
→ ProjectRepository.GetProjectByID
→ ownership check
→ TaskRepository.CreateTask
→ cache invalidation
→ response.SuccessResponse
```

Ví dụ `PATCH /api/v1/tasks/:id/assign`:

```text
AuthMiddleware
→ TaskHandler.AssignedTask
→ bind assignee_id
→ TaskService.AssignTask
→ TaskRepository.GetTaskById
→ ProjectRepository.GetProjectByID
→ ownership check
→ UserRepository.GetUserByID
→ TaskRepository.AssignTask
→ enqueue notification job
→ cache invalidation
→ response.SuccessResponse
```

Điểm quan trọng: task assignment hiện đã có tầng queue job cho notification, tức project đã có mầm cho event-driven behavior, dù chưa phải realtime websocket.

### 2.5 Project flow thực tế

Project service đang phân quyền rõ hơn task/user:

- `ListProjectsByOwner` lọc theo ownerID.
- `UpdateProject` và `DeleteProject` kiểm tra `project.OwnerID == currentUserID`.
- `GetProjectByID` thì chỉ validate ID và fetch project.

Tức là ownership enforcement cho mutate là ở service layer, không phải middleware.

### 2.6 User flow thực tế

User service hiện là CRUD thẳng xuống repository, gần như không có authorization ngoài auth middleware.

Điều này tạo ra rủi ro lớn nhất trong hệ thống hiện tại: bất kỳ authenticated user nào cũng có thể gọi list/get/update/delete user khác nếu biết ID/email/fullname.

### 2.7 Response và error flow

Response được chuẩn hóa qua [internal/response/response.go](internal/response/response.go) với `SuccessResponse` và `ErrorResponse`.

Handler thường map error service sang HTTP status bằng các hàm như `mapTaskErrorToStatus`, `mapUserErrorToStatus`, `mapAuthErrorToStatus`.

Điểm tốt:

- Có một shape response thống nhất.
- Error classification nằm ở handler, nên service không bị dính HTTP.

Điểm chưa tốt:

- Một số handler map lỗi bằng tay, nên logic status code bị phân mảnh theo từng file.
- `UnassignTask` trả entity thô thay vì DTO mapper, làm response contract không đồng nhất.

## 3. Đánh Giá Khả Năng Tích Hợp Realtime

### Chỗ nên trigger realtime event

Các điểm trigger hợp lý nhất là nơi state thay đổi thật sự đã được commit:

- `TaskService.CreateTask`
- `TaskService.UpdateTask`
- `TaskService.AssignTask`
- `TaskService.UnassignTask`
- `TaskService.DeleteTask`
- `ProjectService.UpdateProject`
- `ProjectService.DeleteProject`
- `UserService.UpdateUser` nếu sau này có profile update realtime

Lý do: đây là nơi business state thay đổi có ý nghĩa, nên event phát ra ở đây mới đúng domain semantics.

### Service nào phù hợp để emit event

TaskService là candidate tốt nhất.

Lý do:

- Task là entity có activity cao nhất.
- Task đã có ownership, assignee, cache invalidation, và queue notification.
- Event realtime tự nhiên nhất là task created / updated / assigned / unassigned / deleted.

ProjectService cũng hợp lý để emit event khi project đổi tên, đổi mô tả, hoặc bị xóa.

UserService thì ít phù hợp hơn, trừ khi app thật sự có activity feed hoặc presence.

### Có đang coupling quá mạnh không

Có, ở mức vừa phải.

Hiện TaskService đã biết quá nhiều thứ cùng lúc:

- business validation
- ownership enforcement
- cache invalidation
- queue notification
- repository orchestration

Nếu nhét thêm websocket hub trực tiếp vào service, service sẽ phình thêm trách nhiệm transport / delivery. Đó là coupling xấu cho production.

### Có nguy cơ circular dependency không

Có nguy cơ nếu đi theo hướng naive.

Ví dụ xấu:

- websocket hub import service để query data
- service import hub để broadcast event

Khi đó dễ tạo vòng phụ thuộc giữa application và delivery layer.

Thiết kế an toàn hơn là:

- service chỉ phát domain/application event qua interface
- realtime layer subscribe hoặc consume event đó
- realtime layer tự query read model hoặc nhận payload đã chuẩn hóa

### Có nên dùng event bus nội bộ không

Có, nếu mục tiêu là production-ready và muốn tách task/project mutation khỏi realtime delivery.

Ưu điểm:

- Giảm coupling giữa service và websocket.
- Dễ thêm worker, notification, audit log, analytics sau này.
- Có thể fan-out một event sang nhiều consumer.

Nhược điểm:

- Tăng độ phức tạp.
- Cần xử lý ordering, retry, duplicate, và idempotency.

Nếu system chưa lớn, in-memory event dispatcher là đủ cho phase đầu. Nếu muốn scale nhiều instance, nên đi thẳng sang Redis Pub/Sub hoặc message broker.

### Có nên inject websocket hub vào service không

Không nên là mặc định.

Lý do:

- Service sẽ bị phụ thuộc vào transport layer.
- Test service khó hơn vì phải mock websocket side effects.
- Scale ngang nhiều instance sẽ khó hơn vì hub chỉ sống trong process memory.

Nếu bắt buộc demo nhanh, có thể inject một interface broadcaster rất mỏng. Nhưng với production thinking, tốt hơn là service phát event vào abstraction, còn realtime adapter xử lý broadcast.

## 4. Đề Xuất Kiến Trúc Realtime Phù Hợp

### Folder structure đề xuất

Một layout hợp lý cho codebase hiện tại là:

```text
internal/
  realtime/
    hub/
    connection/
    auth/
    event/
    broadcaster/
    subscriber/
    serializer/
  events/
    task/
    project/
    user/
  application/
    ports/
```

Hoặc nếu muốn ít folder hơn:

```text
internal/realtime
internal/events
internal/application/ports
```

Điểm chính là tách rõ 3 thứ:

- domain/application event
- realtime delivery adapter
- transport / connection management

### Realtime layer placement

Realtime layer nên nằm song song với HTTP layer, không nằm bên trong handler.

Tức là:

- HTTP request vẫn đi qua Gin handler.
- Sau khi service commit xong, nó phát event qua interface.
- Realtime adapter nhận event rồi broadcast.

### Hub / Connection Manager placement

Hub nên là infrastructure component riêng, không nằm trong service.

Hub nên chịu trách nhiệm:

- register/unregister connection
- room membership
- broadcast to room
- backpressure policy
- ping/pong heartbeat
- graceful close

### Event model

Nên dùng event object mang nghĩa nghiệp vụ, không chỉ là raw JSON.

Ví dụ conceptually:

- `TaskCreated`
- `TaskUpdated`
- `TaskAssigned`
- `TaskUnassigned`
- `TaskDeleted`
- `ProjectUpdated`
- `ProjectDeleted`

Mỗi event nên có:

- event id
- type
- occurred at
- actor user id
- project id
- task id nếu có
- payload đã chuẩn hóa
- version

### Broadcast strategy

Ưu tiên theo project room.

Lý do:

- task/project activity là theo project boundary.
- user-specific fan-out chỉ là một lớp phụ, không nên là primary broadcast path.

Chiến lược tốt thường là:

1. Broadcast cho project room.
2. Nếu cần, mirror sang user room cho assignee / owner.
3. Nếu cần audit feed, đẩy sang separate consumer.

### Authentication strategy

WebSocket auth nên reuse JWT hiện tại nhưng không reuse nguyên xi middleware HTTP.

Nên làm:

- validate token ngay lúc handshake hoặc upgrade.
- extract user id từ claims.
- verify user still exists / still active.
- attach authenticated principal vào connection.

Không nên:

- rely vào query params đơn giản.
- để unauthenticated socket tự subscribe room bằng project id.

### Project room strategy

Room model nên xoay quanh authorization boundary:

- project room: tất cả user có quyền xem project đó.
- user room: chỉ dành cho thông báo cá nhân.

Khi socket join project room, cần check:

- owner
- assignee của task trong project
- hoặc explicit membership nếu sau này có team model

### Future scalability strategy

Nếu chỉ một instance:

- in-memory hub là đủ.

Nếu nhiều instance:

- dùng Redis Pub/Sub hoặc stream để fan-out cross-instance.
- mỗi instance giữ local hub.
- mỗi instance subscribe shared topic rồi broadcast sang client local.

Nếu workload tăng mạnh hơn:

- event bus tách biệt khỏi websocket.
- task mutation tạo event.
- separate worker xử lý realtime, notification, audit.

## 5. Production Concerns

### Concurrency

WebSocket hub và connection map phải thread-safe.

Nguy cơ hiện tại nếu thiết kế vội:

- race condition khi register/unregister connection.
- concurrent write vào một socket.

### Goroutine leak

Mỗi connection thường kéo theo goroutine đọc / write.

Nếu không chốt lifecycle rõ, goroutine sẽ leak khi client disconnect hoặc context bị cancel.

### WebSocket disconnect

Phải phát hiện disconnect nhanh và cleanup ngay:

- close read loop
- remove from room
- close write queue
- free buffer

### Ping/pong heartbeat

Nên có heartbeat định kỳ để phân biệt client chết thật và kết nối chỉ bị treo.

Không có heartbeat thì room membership và memory usage sẽ phình dần.

### Backpressure

Khi broadcast nhanh hơn tốc độ đọc của client, queue sẽ phình.

Phải có chính sách:

- bounded buffer
- drop oldest hoặc drop slow client
- ưu tiên message quan trọng

### Slow clients

Slow client không nên block toàn bộ room.

Một client quá chậm phải bị cô lập, nếu không sẽ ảnh hưởng latency của cả hệ thống.

### Message buffering

Buffer cần có giới hạn rõ.

Không nên dùng channel không giới hạn hoặc slice append vô hạn cho mỗi connection.

### Redis Pub/Sub scaling

Redis Pub/Sub hợp cho realtime fan-out đơn giản, nhưng:

- không durable.
- message có thể mất nếu subscriber down.
- không có replay native.

Nếu yêu cầu không được mất event, Redis Streams hoặc broker khác sẽ phù hợp hơn.

### Multiple instances

Nếu chạy nhiều API instance:

- HTTP mutation có thể vào instance A.
- WebSocket client đang nối instance B.

Do đó broadcast không thể chỉ dựa vào memory local.

### Memory leak

Các nguồn leak thường gặp:

- map connection không cleanup.
- room membership không remove.
- queued payload không drain.
- goroutine retry / watchdog không stop.

### Graceful shutdown

Khi app dừng phải:

- stop accept socket mới
- close hub
- drain queue nếu cần
- close current connections có lý do rõ ràng
- chờ worker goroutine thoát an toàn

## 6. Security Review

### JWT validation cho websocket

JWT cho WebSocket cần validate nghiêm như HTTP auth, tối thiểu gồm:

- signature
- expiry
- signing method
- subject / user id hợp lệ
- user tồn tại

Không nên tin token chỉ vì parse được.

### Origin checking

Với browser-based WebSocket client, phải kiểm tra Origin để giảm rủi ro cross-site connection abuse.

### Unauthorized subscription

Không được cho client join room chỉ bằng project id truyền lên.

Phải verify user có quyền với project đó trước khi cho subscribe.

### Project isolation

Project event chỉ nên đến các user có liên quan:

- owner
- assignee
- future project members

Không nên broadcast toàn hệ thống nếu dữ liệu mang tính riêng tư.

### Event spoofing

Client không được phép tự gửi event mang nghĩa server-side.

Chỉ server mới tạo authoritative event.

### Sensitive data leakage

Không nên phát raw entity nếu entity chứa dữ liệu nhạy cảm.

Ví dụ:

- password hash
- email nếu project room không cần thiết
- internal metadata

Nên có event DTO riêng, tối thiểu hóa payload.

## 7. Refactor Recommendation Trước Khi Làm Realtime

### Chỗ cần refactor

- Tách responsibility của TaskService: hiện vừa validate, vừa ownership check, vừa cache, vừa queue.
- Chuẩn hóa response mapping để mọi handler trả DTO thống nhất.
- Bổ sung context.Context cho repository để cancel / timeout propagation tốt hơn.
- Làm JWT config động hơn, không nên gắn chặt ở package init time.
- Làm recovery middleware an toàn hơn với mọi kiểu panic.

### Chỗ coupling chưa tốt

- AuthMiddleware đang phụ thuộc repository và DB lookup mỗi request.
- User CRUD đang thiếu policy layer, nên security logic chưa nằm ở một nơi rõ.
- TaskService đang biết cả queue notification và cache invalidation.

### Service nào nên tách interface

Nên ưu tiên tách interface cho:

- TaskService nếu muốn test và future event emission dễ hơn.
- ProjectService nếu muốn policy / permission / event orchestration rõ hơn.
- AuthService hiện đã có interface nhưng vẫn đang phụ thuộc vào repo cụ thể qua constructor; nên nhất quán với các service khác nếu tiếp tục mở rộng.

### Middleware nào nên cải thiện

- AuthMiddleware: nên giảm việc query DB trong middleware nếu có thể cache principal hoặc chỉ verify token + user active check tối thiểu.
- RecoveryMiddleware: phải an toàn với `error` và mọi kiểu panic.
- LoggerMiddleware: nên log structured hơn, có request id, latency, route template.

### Naming nào chưa chuẩn

- `AssignedTask` nên thống nhất tên là `AssignTask` để khớp ngữ nghĩa route và handler chuẩn.
- `GetTaskById` nên thống nhất style với `GetUserByID` hoặc toàn bộ codebase theo một convention duy nhất.
- `GetAllTasksByUserID` là tên tốt hơn nếu muốn nhấn ownership scope; hiện tại semantics khá rõ, chỉ cần nhất quán.

### Dependency injection nào cần sửa

- [internal/app/app.go](internal/app/app.go) đang là nơi new toàn bộ dependency. Đây là đúng với composition root, nhưng nếu thêm realtime nên kéo event bus / broadcaster / hub vào đây theo interface rõ ràng.
- Không nên inject websocket hub thẳng vào service.
- Nên inject một domain event publisher abstraction thay vì transport cụ thể.

## 8. Kế Hoạch Implementation Từng Bước

### Phase 1

Mục tiêu:

- Chuẩn hóa boundary hiện tại trước khi thêm realtime.

Lý do:

- Nếu kiến trúc lõi còn coupling mạnh, realtime sẽ khuếch đại nợ kỹ thuật rất nhanh.

Expected output:

- service boundary rõ hơn
- response contract thống nhất
- auth / ownership rules dễ đọc hơn

Risk:

- dễ scope creep nếu cố refactor toàn bộ cùng lúc.

Best practice:

- chỉ tách những điểm ảnh hưởng trực tiếp đến realtime event flow.

### Phase 2

Mục tiêu:

- Thêm domain/application event interface ở service layer.

Lý do:

- để task/project mutation có thể phát event mà không biết transport phía sau.

Expected output:

- task/project service emit event qua interface
- consumer chưa cần websocket vẫn có thể log / test / queue

Risk:

- duplicate event hoặc event không đồng bộ nếu không định nghĩa transaction boundary rõ.

Best practice:

- phát event sau khi mutation thành công.
- payload nhỏ, versioned, idempotent.

### Phase 3

Mục tiêu:

- Xây realtime adapter và hub in-memory cho single instance.

Lý do:

- đây là cách nhanh nhất để xác thực flow end-to-end mà chưa cần scale ngang.

Expected output:

- socket connect / disconnect
- room join / leave
- broadcast theo project
- auth handshake bằng JWT

Risk:

- memory leak nếu lifecycle socket chưa kín.

Best practice:

- có heartbeat.
- bounded buffer.
- cleanup chắc chắn.

### Phase 4

Mục tiêu:

- Nâng lên multi-instance distribution.

Lý do:

- realtime single instance sẽ chạm trần rất nhanh nếu API scale.

Expected output:

- Redis Pub/Sub hoặc Streams bridge
- cross-instance broadcast
- local hub vẫn giữ client connections

Risk:

- message ordering, duplicate, and lost message semantics.

Best practice:

- định nghĩa rõ loại event nào best-effort và loại nào phải reliable.

### Phase 5

Mục tiêu:

- Hoàn thiện production hardening.

Lý do:

- realtime hệ thống chỉ thực sự production-ready khi chịu được disconnect, retry, slow clients và shutdown.

Expected output:

- graceful shutdown
- observability tốt hơn
- rate limiting / origin check / auth hardening

Risk:

- nếu không có metrics, rất khó debug lỗi realtime trong môi trường thật.

Best practice:

- log theo request id / connection id / room id.
- đo số connection, buffer depth, drop count, reconnect rate.

---

Kết luận ngắn: codebase hiện tại đủ nền để đi tiếp sang realtime, nhưng chưa nên gắn WebSocket trực tiếp vào service. Hướng an toàn hơn là chuẩn hóa boundary trước, rồi thêm event abstraction, sau đó mới dựng hub và broadcast layer.