

BEGIN;

TRUNCATE TABLE tasks, projects, users
RESTART IDENTITY CASCADE;

INSERT INTO users (full_name, email, password_hash, created_at)
VALUES
('Nguyen Khang', 'khang@gmail.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Alice Nguyen', 'alice.nguyen@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Bob Tran', 'bob.tran@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Linh Pham', 'linh.pham@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Minh Le', 'minh.le@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Hoa Do', 'hoa.do@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW());

INSERT INTO projects (name, description, owner_id, created_at)
VALUES
('Khang Project 1', 'Backend auth module', 1, NOW()),
('Khang Project 2', 'Task API module', 1, NOW()),
('Khang Project 3', 'Project management module', 1, NOW()),

('Alice Project 1', 'Dashboard module', 2, NOW()),
('Alice Project 2', 'Reporting module', 2, NOW()),
('Alice Project 3', 'Notification module', 2, NOW()),

('Bob Project 1', 'Mobile sync module', 3, NOW()),
('Bob Project 2', 'API gateway module', 3, NOW()),
('Bob Project 3', 'File upload module', 3, NOW()),

('Linh Project 1', 'QA automation module', 4, NOW()),
('Linh Project 2', 'Testing dashboard module', 4, NOW()),
('Linh Project 3', 'Bug tracking module', 4, NOW()),

('Minh Project 1', 'Admin portal module', 5, NOW()),
('Minh Project 2', 'Role permission module', 5, NOW()),
('Minh Project 3', 'User audit module', 5, NOW()),

('Hoa Project 1', 'Data cleanup module', 6, NOW()),
('Hoa Project 2', 'Backup module', 6, NOW()),
('Hoa Project 3', 'Analytics module', 6, NOW());
INSERT INTO tasks
(project_id, title, description, status, assignee_id, created_at)
VALUES
(1, 'Setup login API', 'Create login endpoint', 'DONE', 1, NOW()),
(1, 'Setup register API', 'Create register endpoint', 'IN_PROGRESS', 2, NOW()),
(1, 'Validate JWT', 'Add auth middleware', 'TODO', 3, NOW()),

(2, 'Create task API', 'Implement create task API', 'DONE', 1, NOW()),
(2, 'Update task API', 'Implement update task API', 'IN_PROGRESS', 4, NOW()),
(2, 'Delete task API', 'Implement delete task API', 'TODO', 5, NOW()),

(7, 'Sync endpoint', 'Implement sync API', 'DONE', 3, NOW()),
(7, 'Delta sync', 'Return changed data only', 'IN_PROGRESS', 1, NOW()),
(7, 'Sync tests', 'Write sync tests', 'TODO', 2, NOW()),

(8, 'Gateway routing', 'Setup route forwarding', 'DONE', 3, NOW()),
(8, 'Gateway auth', 'Validate token at gateway', 'IN_PROGRESS', 4, NOW()),
(8, 'Gateway logging', 'Log gateway requests', 'TODO', 5, NOW()),

(9, 'Upload file', 'Create upload endpoint', 'DONE', 3, NOW()),
(9, 'Validate file', 'Check file type and size', 'IN_PROGRESS', 6, NOW()),
(9, 'Store file URL', 'Save uploaded file URL', 'TODO', 1, NOW()),

(10, 'Smoke tests', 'Create smoke tests', 'DONE', 4, NOW()),
(10, 'Regression tests', 'Create regression tests', 'IN_PROGRESS', 1, NOW()),
(10, 'CI testing', 'Run tests in CI', 'TODO', 2, NOW()),

(11, 'Testing metrics', 'Build metrics API', 'DONE', 4, NOW()),
(11, 'Status board', 'Create testing board', 'IN_PROGRESS', 3, NOW()),
(11, 'Failed logs', 'Store failed logs', 'TODO', 5, NOW()),

(12, 'Create bug', 'Implement create bug API', 'DONE', 4, NOW()),
(12, 'Assign bug', 'Assign bug to developer', 'IN_PROGRESS', 5, NOW()),
(12, 'Close bug', 'Implement close bug flow', 'TODO', 6, NOW()),

(13, 'Admin dashboard', 'Create admin dashboard', 'DONE', 5, NOW()),
(13, 'User management', 'Manage all users', 'IN_PROGRESS', 1, NOW()),
(13, 'Project management', 'Manage all projects', 'TODO', 2, NOW()),

(14, 'Role matrix', 'Define role permissions', 'DONE', 5, NOW()),
(14, 'Role middleware', 'Check user role', 'IN_PROGRESS', 3, NOW()),
(14, 'Permission tests', 'Test forbidden cases', 'TODO', 4, NOW()),

(15, 'Audit event', 'Create audit model', 'DONE', 5, NOW()),
(15, 'Audit history', 'List audit history', 'IN_PROGRESS', 2, NOW()),
(15, 'Audit filter', 'Filter audit logs', 'TODO', 6, NOW()),

(16, 'Find dirty data', 'Detect invalid records', 'DONE', 6, NOW()),
(16, 'Clean old tasks', 'Remove old tasks', 'IN_PROGRESS', 1, NOW()),
(16, 'Cleanup report', 'Generate cleanup report', 'TODO', 2, NOW()),

(17, 'Backup database', 'Create backup script', 'DONE', 6, NOW()),
(17, 'Restore database', 'Create restore script', 'IN_PROGRESS', 5, NOW()),
(17, 'Backup scheduler', 'Schedule daily backup', 'TODO', 3, NOW()),

(18, 'Analytics summary', 'Create analytics summary', 'DONE', 6, NOW()),
(18, 'Task analytics', 'Analyze task status', 'IN_PROGRESS', 3, NOW()),
(18, 'User analytics', 'Analyze user workload', 'TODO', 4, NOW());

COMMIT;