# PROJECT ANALYSIS

## Scope

Tài liệu này phân tích kiến trúc hiện tại của Go backend dựa trên các thành phần chính:

- `cmd/main.go`
- `internal/app/app.go`
- `internal/routes/*`
- `internal/handler/*`
- `internal/services/*`
- `internal/repositories/*`
- `internal/cache/*`
- `internal/database/*`
- `internal/middleware/*`
- `internal/dto/*`
- `internal/entities/*`
- `internal/mapper/*`
- `internal/validation/*`
- `go.mod`
- `README.md`

Mục tiêu của phân tích là xác định mức độ sẵn sàng để mở rộng sang Notification Job Queue, Worker Process riêng, Retry Logic, WebSocket Server, và Realtime Task Event Broadcasting.

---

## 1. Phân Tích Cấu Trúc Thư Mục

### 1.1 Tổng Quan Hiện Tại

Dự án hiện đang là một backend REST API dạng layered monolith, với các lớp chính:

- Transport layer: `handler/`, `middleware/`, `routes/`
- Application layer: `services/`
- Persistence layer: `repositories/`, `database/`
- Shared model layer: `entities/`, `dto/`, `mapper/`, `validation/`
- Cross-cutting infrastructure: `cache/`, `config/`, `response/`

### 1.2 Ý Nghĩa Từng Folder

### `cmd/`
Entry point của binary. Hiện tại chỉ có một entrypoint ở `cmd/main.go`, và file này gọi vào `internal/app.Run()`. Đây là dấu hiệu của một ứng dụng đơn binary HTTP.

### `internal/app/`
Composition root của hệ thống. Nơi khởi tạo PostgreSQL, Redis, repository, service, handler, và routes. Đây là nơi dependency injection đang được làm thủ công.

### `internal/config/`
Load cấu hình từ environment variables. Hiện tại có `DBConfig` và `RedisConfig`.

### `internal/database/`
Chứa factory kết nối hạ tầng: PostgreSQL và Redis. Đây là tầng infrastructure, không phải domain.

### `internal/cache/`
Abstraction cho Redis cache. Đây là một bước đúng hướng vì service không gọi trực tiếp Redis client mà đi qua interface `Cache`.

### `internal/dto/`
Chứa request/response contract cho HTTP. Folder này tách biệt dữ liệu API khỏi entity nội bộ, phù hợp với Clean Architecture ở mức cơ bản.

### `internal/entities/`
Chứa core data model: `User`, `Project`, `Task`. Các entity này đang đóng vai trò mô hình nghiệp vụ tối thiểu, nhưng chưa có domain behavior phong phú.

### `internal/handler/`
HTTP controllers. Handler nhận request, bind JSON, lấy user context, gọi service, rồi map response.

### `internal/middleware/`
Middleware cho auth, logger, recovery, request ID. Có file `role.go` nhưng hiện chỉ là skeleton/commented-out, chưa được dùng để enforce role-based access.

### `internal/mapper/`
Chuyển đổi giữa request DTO, entity, và response DTO.

### `internal/repositories/`
Tầng truy cập dữ liệu với GORM. Repository đang encapsulate query logic khá tốt, nhưng một số query đã mang ý nghĩa nghiệp vụ như ownership/assignment.

### `internal/routes/`
Định tuyến Gin và gắn middleware. Hiện đang chia thành auth/task/user/project routes.

### `internal/services/`
Business/use-case layer. Đây là nơi kiểm tra quyền sở hữu project/task, validate đầu vào, gọi repository, và xử lý cache invalidation.

### `internal/validation/`
Hàm validate thuần cho id, email, password, status, project fields.

### `migrations/`
SQL migration cho schema và seed data. Đây là nguồn sự thật cho cấu trúc DB.

### `postman/`
Collection và environment để test API.

### `system/redis/` và `redis.conf/`
Tệp cấu hình runtime cho Redis. Có vẻ phục vụ môi trường local/dev hơn là kiến trúc ứng dụng.

### Folder hiện đang thiếu cho mục tiêu mở rộng

Hiện chưa có các folder sau:

- `cmd/worker/`
- `internal/jobs/`
- `internal/worker/`
- `internal/websocket/`
- `internal/realtime/`
- `internal/events/`
- `internal/queue/`

Đây là các phần nên bổ sung trước khi triển khai notification, queue, realtime event, và retry.

---

## 2. Phân Tích Kiến Trúc Hiện Tại

### Có đang theo Clean Architecture không?

Chưa phải Clean Architecture đúng nghĩa. Nó gần với layered architecture hơn:

- `handler` đóng vai delivery layer
- `services` đóng vai application/business layer
- `repositories` đóng vai persistence layer
- `entities`, `dto`, `mapper`, `validation` là shared model/support layer

Điểm đúng:

- Handler không chứa SQL.
- Service không phụ thuộc Gin trực tiếp.
- Repository không biết về HTTP.
- DTO và entity đã tách ra.

Điểm chưa sạch:

- Composition root đang nằm trong `internal/app/app.go`, nhưng mọi dependency đều được new trực tiếp tại đây, chưa có interface cho application boundary.
- Một số nghiệp vụ vẫn bị “chìm” vào repository query, ví dụ `GetAllTasksByUserID` mang semantics ownership/assignment.
- `mapper` tồn tại, nhưng một số mapping/normalization vẫn làm rải rác ở handler/service.

### Dependency flow hiện tại

Luồng phụ thuộc chính đang là:

```text
cmd/main.go
  -> internal/app/app.go
    -> routes
      -> middleware
      -> handler
        -> services
          -> repositories
            -> database / cache
```

Đây là hướng dependency tương đối hợp lý cho một REST API đơn binary.

### Service có phụ thuộc HTTP không?

Không trực tiếp. Service không import Gin hay HTTP primitives.

Tuy nhiên, service đang phụ thuộc vào `context.Context` lấy từ request, và một số use case vẫn gắn chặt vào hành vi HTTP như cache invalidation theo request lifecycle. Điều này chưa sai, nhưng sẽ cần trừu tượng lại khi thêm worker hoặc event-driven flow.

### Repository có phụ thuộc business logic không?

Repository không phụ thuộc service, nhưng đã chứa một phần business semantics:

- `GetAllTasksByUserID()` dùng join để diễn giải quyền truy cập theo owner hoặc assignee.
- `AssignTask()` và `UnassignTask()` không chỉ update data mà còn ngầm giả định workflow nghiệp vụ liên quan đến assignment state.

Tức là repository vẫn còn hơi “nghiêng” về use-case semantics, chưa hoàn toàn chỉ làm persistence.

### DTO / Entity / Mapper đã tách đúng chưa?

Khá đúng ở mức cơ bản:

- DTO dùng cho API boundary.
- Entity dùng cho model nội bộ.
- Mapper chuyển đổi giữa hai phía.

Nhưng vẫn còn một vài điểm cần chú ý:

- Một số request object vẫn được service đọc trực tiếp mà chưa qua một application command object rõ ràng.
- `mapper` đang làm cả defaulting logic, ví dụ task status mặc định `TODO`.

### Cache layer đang ở đâu?

Cache layer nằm ở `internal/cache/` và được inject vào `TaskService`.

Đây là thiết kế tốt hơn việc để service gọi Redis client trực tiếp, vì nó tạo một abstraction để sau này có thể thay cache backend hoặc mở rộng thành queue adapter.

### Redis đang được inject thế nào?

Redis được khởi tạo ở `internal/app/app.go` thông qua `database.ConnectRedis(config.NewRedisConfig())`, sau đó wrap bằng `cache.NewRedisCache(redisClient)` và inject vào `TaskService`.

Điểm tốt:

- Dependency được tạo ở composition root.
- Service chỉ thấy interface `cache.Cache`.

Điểm cần cải thiện:

- Redis vẫn đang là dependency hạ tầng được tạo trực tiếp trong runtime HTTP app.
- Chưa có abstraction cho queue/stream/pubsub, nên hiện tại Redis mới chỉ là cache.

### Context propagation đã đúng chưa?

Chỉ đúng một phần.

Hiện tại handler lấy `c.Request.Context()` và truyền xuống service/cache. Đây là đúng hướng.

Nhưng repository hiện không nhận `context.Context`, nên các truy vấn DB không thể tận dụng cancellation/deadline từ request. Khi thêm worker, job queue, hoặc retry, đây sẽ là điểm nghẽn rõ ràng.

Ngoài ra, middleware auth dựa vào Gin context để nhét `user_id` và `current_user`, nên phần context propagation đang chia đôi giữa `gin.Context` và `context.Context` mà chưa có chiến lược thống nhất.

---

## 3. Phân Tích Flow Request Hiện Tại

### 3.1 Login Flow

```text
Client
  -> POST /api/v1/auth/login
  -> AuthHandler.Login
  -> AuthService.Login
  -> UserRepository.GetUserByEmail
  -> PostgreSQL
  -> bcrypt password compare
  -> utils.GenerateAccessToken
  -> JSON response
```

Điểm lưu ý:

- Login không đi qua auth middleware.
- Token được sinh ở service layer.
- Secret JWT được đọc từ environment ở package init time trong `utils/jwt.go`.

### 3.2 Create Task Flow

```text
Client
  -> POST /api/v1/tasks
  -> AuthMiddleware
  -> TaskHandler.CreateTask
  -> mapper.CreateTaskRequestToTaskEntity
  -> TaskService.CreateTask
  -> ProjectRepository.GetProjectByID
  -> TaskRepository.CreateTask
  -> cache.Delete(task list key)
  -> JSON response
```

Đặc điểm:

- Handler làm request binding và lấy current user.
- Service kiểm tra project ownership.
- Task status được chuẩn hóa sang uppercase và validate.
- Cache của danh sách task theo user bị invalidate sau khi tạo mới.

### 3.3 Assign Task Flow

```text
Client
  -> PATCH /api/v1/tasks/:id/assign
  -> AuthMiddleware
  -> TaskHandler.AssignedTask
  -> TaskService.AssignTask
  -> TaskRepository.GetTaskById
  -> ProjectRepository.GetProjectByID
  -> UserRepository.GetUserByID
  -> TaskRepository.AssignTask
  -> cache.Delete(task key)
  -> cache.Delete(owner task list key)
  -> cache.Delete(assignee task list key)
  -> JSON response
```

Đây là flow có nhiều side effect nhất trong service hiện tại.

### 3.4 Cache Invalidation Flow

```text
Read path
  -> cache.Get
  -> cache hit: decode JSON and return
  -> cache miss: query DB, marshal, cache.Set with TTL 5 minutes

Write path
  -> create/update/delete/assign/unassign
  -> cache.Delete on task key and user task list keys
```

Nhận xét:

- Read-through cache có tồn tại cho task detail và task list theo user.
- Invalidation đang làm thủ công theo từng use case.
- TTL cố định 5 phút.

---

## 4. Phân Tích Khả Năng Mở Rộng

### Worker Process riêng

Chưa phù hợp ngay nếu giữ nguyên cấu trúc hiện tại.

Lý do:

- Chưa có binary riêng cho worker.
- Chưa có job abstraction.
- Service hiện đang xử lý synchronous side effects trực tiếp.
- Repository và service chưa được thiết kế để chạy ngoài HTTP lifecycle.

### Redis Queue

Có thể thêm, nhưng nên refactor trước.

Lý do:

- Redis hiện chỉ là cache wrapper.
- Chưa có queue/stream interface để producer và consumer cùng nói chuyện qua một contract.
- Chưa có retry metadata, dead-letter strategy, hoặc job status tracking.

### Notification Job

Chưa sẵn sàng.

Lý do:

- Hiện tại task create/assign/update chưa phát event domain rõ ràng.
- Không có event bus hoặc outbox pattern.
- Nếu thêm notification trực tiếp vào service, service sẽ bị phình ra và khó test.

### WebSocket

Chưa sẵn sàng về mặt kiến trúc.

Lý do:

- Không có hub quản lý connection.
- Không có lifecycle cleanup cho connection.
- Không có broadcast layer tách khỏi HTTP handlers.

### Realtime Event Broadcasting

Có thể làm, nhưng chỉ sau khi tách event boundary.

Lý do:

- Hiện tại use case chỉ trả response HTTP, chưa phát domain event.
- Không có subscriber/projection layer.
- Không có event naming/versioning contract.

### Retry Mechanism

Chưa có chỗ đặt retry policy đúng nghĩa.

Lý do:

- Chưa có worker loop.
- Chưa có queue state hay attempt count.
- Chưa có idempotency key.

### Kết luận khả năng mở rộng

Project hiện phù hợp cho CRUD REST API, nhưng để đi sang worker + realtime thì cần thêm ít nhất một lớp event/queue boundary. Nếu không, side effects sẽ tiếp tục bị gắn chặt trong service layer và rất khó scale.

---

## 5. Đề Xuất Cấu Trúc Mới

### Mục tiêu cấu trúc

- Tách API process và worker process thành hai binary riêng.
- Tách synchronous use case và asynchronous job handling.
- Có event layer để broadcast realtime và enqueue notification.
- Giữ dependency hướng vào trong, tránh vòng phụ thuộc giữa HTTP, worker, websocket, queue.

### Cấu trúc đề xuất

```text
cmd/
  api/
    main.go
  worker/
    main.go

internal/
  app/
  config/
  database/
  cache/
  dto/
  entities/
  mapper/
  validation/
  response/
  middleware/
  handler/
  routes/
  services/
  repositories/

  queue/
    producer.go
    consumer.go
    message.go
    redis_queue.go

  jobs/
    notification_job.go
    task_event_job.go
    retry_job.go

  worker/
    runner.go
    processor.go
    shutdown.go

  websocket/
    hub.go
    client.go
    upgrader.go
    broadcaster.go

  realtime/
    service.go
    dispatcher.go
    subscribers.go

  events/
    task_created.go
    task_updated.go
    task_assigned.go
    task_unassigned.go
```

### Gợi ý thêm về ranh giới trách nhiệm

- `queue/` chỉ lo enqueue/dequeue và envelope format.
- `jobs/` chỉ định nghĩa payload và handler job.
- `worker/` chỉ lo poll, execute, retry, shutdown.
- `websocket/` chỉ lo connection management và broadcast.
- `realtime/` chỉ lo chuyển domain event sang notification/broadcast.
- `events/` chỉ chứa event contract và metadata.

---

## 6. Phân Tích Điểm Nguy Hiểm Tiềm Ẩn

### Race condition

Hiện chưa thấy shared in-memory state lớn, nên rủi ro race condition thấp ở runtime HTTP hiện tại.

Tuy nhiên, khi thêm WebSocket hub hoặc worker queue, shared maps/channels sẽ cần mutex hoặc single-threaded event loop rõ ràng.

### Stale cache

Rủi ro hiện hữu.

Nguyên nhân:

- Cache invalidation đang làm thủ công.
- Chỉ xóa một số key liên quan đến user/task hiện tại.
- Không có invalidate theo project scope hoặc event scope.

### Duplicated business logic

Rủi ro trung bình.

Nguyên nhân:

- Validation, authorization, and ownership checks xuất hiện trong nhiều service khác nhau.
- Một số mapping/defaulting logic nằm cả ở mapper lẫn service.

### Repository leaking business rules

Rủi ro có thật.

Nguyên nhân:

- Repository `GetAllTasksByUserID()` chứa semantics ownership/assignee.
- Repository đang gánh query mang nghĩa nghiệp vụ thay vì chỉ fetch data thuần.

### Circular dependency risk

Rủi ro tăng mạnh nếu thêm realtime/worker mà không có event boundary.

Nguyên nhân:

- WebSocket/broadcast layer có thể bị gọi ngược vào service.
- Worker nếu import handler hoặc routes sẽ tạo phụ thuộc ngược không cần thiết.

### WebSocket memory leak

Chưa xảy ra vì chưa có WebSocket server, nhưng sẽ là rủi ro lớn nếu không có:

- connection registry
- unregister on close
- heartbeat/ping-pong
- backpressure policy

### Goroutine leak

Rất dễ phát sinh khi thêm worker hoặc broadcaster.

Nguyên nhân:

- Chưa có context cancellation chiến lược.
- Chưa có shutdown path cho long-running loop.
- Chưa có timeout/deadline propagation xuống DB/queue.

### Missing context propagation

Rủi ro hiện tại.

Nguyên nhân:

- Handler có request context, nhưng repository không nhận context.
- DB query không thể hủy khi request bị drop.

### Cache inconsistency

Rủi ro cao hơn khi có nhiều writer.

Nguyên nhân:

- Cache invalidation không theo event.
- Không có outbox/retry để đảm bảo event và write nhất quán.

### Retry infinite loop risk

Nếu thêm retry mà không có max attempts, dead-letter queue, hoặc backoff, job sẽ dễ lặp vô hạn.

---

## 7. Đề Xuất Roadmap Implementation

### Bước 1. Refactor Redis layer

Mục tiêu:

- Tách cache client ra khỏi ý nghĩa storage tạm thời.
- Chuẩn hóa interface để sau này dùng chung cho cache và queue.

Files cần tạo/sửa:

- Sửa `internal/cache/*`
- Sửa `internal/app/app.go`
- Sửa `internal/database/redis.go`

Dependency ảnh hưởng:

- `TaskService`
- mọi nơi đang gọi cache

Risk cần chú ý:

- Không làm vỡ cache key hiện tại.
- Không trộn cache API với queue API ngay từ đầu.

### Bước 2. Build queue abstraction

Mục tiêu:

- Tạo interface producer/consumer/job envelope.
- Cho phép enqueue task events và notification jobs.

Files cần tạo/sửa:

- Tạo `internal/queue/*`
- Tạo `internal/jobs/*`
- Sửa `go.mod` nếu dùng Redis stream/list library mới

Dependency ảnh hưởng:

- `TaskService`
- `Notification` future flow

Risk cần chú ý:

- Idempotency.
- Message schema versioning.

### Bước 3. Build notification producer

Mục tiêu:

- Phát event khi task được tạo, cập nhật, assign, unassign.
- Producer chỉ đẩy message, không xử lý notification trực tiếp.

Files cần tạo/sửa:

- Tạo `internal/events/*`
- Sửa `internal/services/task_service.go`
- Sửa `internal/services/project_service.go` nếu cần event liên quan project

Dependency ảnh hưởng:

- `TaskService`
- `queue/`

Risk cần chú ý:

- Không publish event trước khi DB commit nếu chưa có transaction/outbox.

### Bước 4. Build worker process

Mục tiêu:

- Chạy consumer riêng để xử lý job async.
- Tách process khỏi API server.

Files cần tạo/sửa:

- Tạo `cmd/worker/main.go`
- Tạo `internal/worker/*`

Dependency ảnh hưởng:

- `queue/`
- `jobs/`
- `database/`

Risk cần chú ý:

- Shutdown an toàn.
- Retry policy.
- Monitoring/logging.

### Bước 5. Add retry mechanism

Mục tiêu:

- Retry có kiểm soát với backoff và max attempts.
- Tránh job chết im lặng hoặc lặp vô hạn.

Files cần tạo/sửa:

- Tạo `internal/jobs/retry_*`
- Tạo `internal/worker/retry_policy.go`

Dependency ảnh hưởng:

- `worker/`
- `queue/`

Risk cần chú ý:

- Dead-letter queue.
- Idempotent side effects.

### Bước 6. Build WebSocket hub

Mục tiêu:

- Quản lý connection lifecycle.
- Broadcast task events realtime.

Files cần tạo/sửa:

- Tạo `internal/websocket/*`
- Sửa routing hoặc app setup để expose WS endpoint

Dependency ảnh hưởng:

- `realtime/`
- `events/`

Risk cần chú ý:

- Memory leak.
- Backpressure.
- Disconnect cleanup.

### Bước 7. Broadcast realtime event

Mục tiêu:

- Chuyển domain event thành realtime message cho client.

Files cần tạo/sửa:

- Tạo `internal/realtime/*`
- Sửa `TaskService` và worker consumer để emit event

Dependency ảnh hưởng:

- `websocket/`
- `queue/`
- `events/`

Risk cần chú ý:

- Duplicate broadcast.
- Event ordering.
- Client-side dedup.

---

## 8. Recommendation

### Nên giữ

- Tách `handler`, `service`, `repository` như hiện tại.
- Giữ `cache.Cache` interface thay vì gọi Redis trực tiếp.
- Giữ request/response DTO tách khỏi entity.

### Nên refactor sớm

- Bổ sung context vào repository.
- Tách event boundary khỏi service.
- Tách API process và worker process.
- Chuẩn hóa cache invalidation theo event thay vì delete thủ công rải rác.

### Nên tránh khi triển khai feature mới

- Không đưa WebSocket logic vào handler HTTP hiện tại.
- Không để service vừa làm business logic vừa chạy worker loop.
- Không để repository chứa thêm rule nghiệp vụ mới.

---

## Warnings

### 1. Cấu hình môi trường có dấu hiệu lỗi

Hàm `GetEnvAsInt()` trong `internal/utils/helper.go` đang có logic đảo ngược: nếu biến môi trường tồn tại thì lại trả default, còn khi rỗng mới thử parse. Điều này có thể làm cấu hình Redis/DB không hoạt động như mong đợi.

### 2. JWT secret được đọc quá sớm

`internal/utils/jwt.go` đọc `JWT_SECRET` ở package init time. Nếu env chưa có lúc binary khởi tạo package, token signing/validation sẽ dùng secret rỗng.

### 3. Recovery middleware chưa an toàn với mọi kiểu panic

`internal/middleware/recovery.go` đang cast panic thành string trực tiếp. Nếu panic không phải string, middleware có thể tự gây lỗi tiếp.

### 4. Authorization role chưa được enforce

Có file `internal/middleware/role.go`, nhưng hiện chỉ là skeleton/commented-out và chưa được gắn vào routes. Điều này đồng nghĩa protected routes hiện mới là authenticated-only, chưa phải role-aware.

### 5. Cache invalidation chưa event-driven

Khi có nhiều writer hoặc sau này thêm worker, invalidate thủ công sẽ khó đảm bảo nhất quán.

---

## Conclusion

Kiến trúc hiện tại phù hợp với REST CRUD monolith và đã có nền tảng tốt ở các phần handler/service/repository/DTO/entity/cache. Tuy nhiên, để thêm Notification Job Queue, Worker Process, Retry Logic, WebSocket Server, và Realtime Event Broadcasting, dự án cần một lớp event/queue boundary rõ ràng hơn, cộng với binary worker riêng và context propagation đầy đủ xuống persistence layer.

Nếu giữ nguyên cấu trúc hiện tại rồi thêm realtime/async feature trực tiếp vào service, hệ thống sẽ nhanh chóng bị coupling cao, cache inconsistency, và khó kiểm soát retry/shutdown.