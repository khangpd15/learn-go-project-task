

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
('Alice Project 3', 'Notification module', 2, NOW());

INSERT INTO tasks
(project_id, title, description, status, assignee_id, created_at)
VALUES
(1, 'Setup login API', 'Create login endpoint', 'DONE', 1, NOW()),
(1, 'Setup register API', 'Create register endpoint', 'IN_PROGRESS', 2, NOW()),
(1, 'Validate JWT', 'Add auth middleware', 'TODO', 3, NOW()),

(2, 'Create task API', 'Implement create task API', 'DONE', 1, NOW()),
(2, 'Update task API', 'Implement update task API', 'IN_PROGRESS', 4, NOW()),
(2, 'Delete task API', 'Implement delete task API', 'TODO', 5, NOW());

COMMIT;