# Backend Code Review — Week 1 & Week 2

## 1. Executive Summary
- Coverage: ~78% of requested Week‑1/Week‑2 items implemented. Core domains (User/Project/Task), DB migrations, Docker compose, JWT auth, bcrypt, and service unit tests exist.
- Production-ready? PARTIAL — functional base present but requires fixes before submission: hardcoded DB credentials, inconsistent error handling, mapper maps plaintext password into `PasswordHash`, middleware returns raw `gin.H`, missing Project Create and GetCurrentUser endpoints.
- Strengths: Clear package separation, DTOs + mappers, GORM repositories, unit tests for TaskService and AuthService, migration files with up/down and constraints.
- Main weaknesses: security-sensitive hardcoding, inconsistent error mapping and sentinel errors, handler/middleware response inconsistencies, missing `category` module.

## 2. Week 1 Checklist
| Task ID | Requirement | Status | Evidence | Comment |
|---|---:|---|---|---|
| W1-T01 | Setup Go project, module, structure, formatter/linter | PARTIAL | `go.mod` present ([go.mod](go.mod#L1)) | Project layout OK; no linter config found. |
| W1-T02 | Define Task domain model | PASS | `internal/entities/task.go` | `Task` struct and `NewTask()` exist. |
| W1-T03 | Repository layer (CRUD) | PASS | `internal/repositories/task_repository.go` | GORM repo implements GetAll/GetById/Create/Update/Delete. |
| W1-T04 | Task service, input/status validation | PASS | `internal/services/task_service.go`, `internal/validation/task_validation.go` | Service enforces validations and ownership checks. |
| W1-T05 | REST API CRUD for Task | PASS | `internal/handler/task_handler.go` | All CRUD handlers implemented; see note about error comparison bug. |
| W1-T06 | Standardize response format | PARTIAL | `internal/response/response.go` used by handlers | Handlers use wrapper but `middleware/auth.go` returns raw `gin.H`. |
| W1-T07 | Middleware: logging, recovery, request ID | PASS | `internal/middleware/logger.go`, `request_id.go`, `recovery.go` and `cmd/main.go` | Wired in `cmd/main.go`. |
| W1-T08 | Unit tests for service create/update/delete task | PASS | `internal/services/task_service_test.go` | Tests use mocks and cover create/update/delete. |
| W1-T09 | Category module assessment (CRUD) | FAIL | No `category` package found | Search returned no results; module missing. |

## 3. Week 2 Checklist
| Task ID | Requirement | Status | Evidence | Comment |
|---|---:|---|---|---|
| W2-T01 | Setup Postgres/MySQL via Docker Compose | PASS | `docker-compose.yml` | Postgres service defined; `.env` used. |
| W2-T02 | Design schema users/projects/tasks | PASS | `migrations/000001_create_users_projects_tasks.up.sql` | Tables and FKs present. |
| W2-T03 | Add migration files | PASS | `migrations/*.up.sql` and `.down.sql` | Multiple migrations and seeds present. |
| W2-T04 | Replace in-memory repo with real DB | PASS (config issue) | `internal/repositories/*.go`, `internal/database/postgres.go` | Repos use GORM; **critical**: DSN hardcoded in `postgres.go`. |
| W2-T05 | Build User module | PASS | `internal/entities/user.go`, `internal/repositories/user_repository.go`, `internal/services/user_service.go` | CRUD and validations present. |
| W2-T06 | Registration API with password hashing | PASS | `internal/services/auth_service.go` | `bcrypt.GenerateFromPassword` used. |
| W2-T07 | Login API returns JWT | PASS | `internal/services/auth_service.go`, `internal/utils/jwt.go` | Token contains `user_id`, `email`, `exp`. |
| W2-T08 | Auth middleware protecting APIs | PASS (style) | `internal/middleware/auth.go`, `routes/routes.go` | Middleware functional; responses use raw `gin.H`. |
| W2-T09 | Project CRUD with owner permission | PARTIAL | `internal/services/project_service.go`, `internal/handler/project_handler.go` | `CreateProject` missing; Update/Delete check owner. |
| W2-T10 | Clean Architecture refactor | PARTIAL | handlers/services/repos/dto/mapper present | Generally separated but mapper/user mapping and some handler logic mix exist. |

## 4. Architecture Review
- Overall separation: handlers → services → repositories is respected. Examples:
  - Handler delegates: `internal/handler/task_handler.go` → `TaskService`.
  - Service business rules: `internal/services/task_service.go` contains ownership and status validation.
  - Repository DB access: `internal/repositories/*.go` use GORM queries only.
- DTO/mapper: `internal/mapper/*` exist and are used.
- Violations and issues:
  - `internal/mapper/user_mapper.go` maps plaintext DTO password to `PasswordHash` field — semantic leak and risk.
  - Some handlers return entity directly (e.g., `CreateTask` returns created entity instead of DTO response mapping).
  - Mixed error signaling (ad‑hoc `errors.New` vs exported sentinel errors) leads to fragile `errors.Is` checks.

## 5. Auth & Security Review
- Register validation: `AuthService.Register` validates fields and email/password first — good (`internal/services/auth_service.go`).
- Password hashing: `bcrypt` used in `AuthService.Register` and `UserService.CreateUser` — good.
- Sensitive logging: no evidence of logging raw passwords or hashes. `logger.go` prints only method/path/status/duration.
- Login error messaging: `AuthService.Login` returns generic `invalid email or password` — correct practice.
- JWT claims: `user_id`, `email`, `exp`, `iat` included in `internal/utils/jwt.go`.
- Auth middleware: validates Bearer token, sets `user_id` and `current_user` in context — functional, but returns raw `gin.H` errors (inconsistent format).
- Secrets/config: `.env` contains `JWT_SECRET` and DB creds but `internal/database/postgres.go` uses hardcoded DSN — critical fix required.

## 6. Authorization Review
- Project ownership: `ProjectService.UpdateProject` and `DeleteProject` enforce owner check — good.
- Task ownership: `TaskService` checks project owner before create/update/delete — good.
- User endpoints exposure: `GET /api/v1/users` is available to any authenticated user; may be overly permissive depending on spec.
- Missing : `GetCurrentUser` (endpoint returning the current authenticated user's profile) is not found.

## 7. Database & Migration Review
- Tables: `users`, `projects`, `tasks` exist in migrations (`migrations/000001_create_users_projects_tasks.up.sql`).
- Additional constraints and indexes applied in `000002_harden_constraints_indexes.up.sql` (status CHECK, indexes, case-insensitive email index).
- ON DELETE behavior: `projects.owner_id` uses `ON DELETE CASCADE`; `tasks.assignee_id` uses `ON DELETE SET NULL`.
- Up/down migrations present. Seed data scripts included.
- Hardcoded DB DSN in `internal/database/postgres.go` is a critical security/config issue.

## 8. API & Error Handling Review
- Task CRUD implemented in `internal/handler/task_handler.go`.
- Project CRUD: Create missing; other verbs implemented in `internal/handler/project_handler.go`.
- User/Auth: `Register`, `Login`, `Logout` implemented; `GetCurrentUser` missing.
- Response format: mostly uses `response.SuccessResponse`/`ErrorResponse`, but `middleware/auth.go` returns raw `gin.H` — unify.
- Error mapping issues:
  - `TaskHandler.GetTaskById` incorrectly compares to `errors.New("forbidden")` (will not match sentinel), causing wrong 404 vs 403.
  - `AuthService.Register` returns ad‑hoc string for duplicate email; `AuthHandler.Register` returns 400 instead of 409 (conflict). Standardize to sentinel and map to 409.

## 9. Unit Test Review
- Unit tests exist for `TaskService` and `AuthService` (`internal/services/*_test.go`).
- Handler tests are commented out (`internal/handler/task_handler_test.go`) — re-enable or replace with proper tests.
- Missing tests: `ProjectService`, `UserService`, `AuthMiddleware`.

## 10. Critical Findings
- Critical
  1. Hardcoded DB credentials in `internal/database/postgres.go`.
  2. `internal/mapper/user_mapper.go` maps plaintext password into `PasswordHash`.
  3. Fragile ad-hoc error strings prevent reliable `errors.Is` matching (e.g., `TaskHandler.GetTaskById`).
- High
  4. Middleware uses raw `gin.H` instead of `ApiResponse` wrapper.
  5. Missing `CreateProject` flow.
- Medium/Low
  6. Over-permissive `GET /api/v1/users`.
  7. Commented handler tests.

## 11. Priority Fix Plan
- Critical fixes first:
  - Replace hardcoded DSN with env-driven config in `internal/database/postgres.go`.
  - Fix `internal/mapper/user_mapper.go` to avoid mapping plaintext into `PasswordHash`.
  - Standardize exported sentinel errors and update handlers to use `errors.Is`.
  - Fix `TaskHandler.GetTaskById` error comparison and `AuthService.Register` to return sentinel `ErrEmailAlreadyExists` mapped to 409.
- High:
  - Normalize middleware responses to use `response.ErrorResponse`.
  - Implement `CreateProject` plus tests.
- Medium/Low:
  - Re-enable handler tests and tighten `GET /api/v1/users` authorization.

## 12. Final Verdict
- Week 1: PARTIAL — core items implemented, but a few correctness/security issues remain.
- Week 2: PARTIAL — DB, migrations and auth implemented but critical config/security and missing features remain.
- Ready to submit? NO — fix Critical and High items first.

---
If you want I can apply the top fixes now: (1) make DB config env-driven, (2) fix `user_mapper` mapping, (3) standardize duplicate-email sentinel and handler mapping, and (4) fix `TaskHandler.GetTaskById`. Reply `apply fixes` to proceed.
