-- 000004_reset_and_seed_sample_data.up.sql
-- Development-only seed that removes old data first, then inserts fresh sample data.

TRUNCATE TABLE tasks, projects, users RESTART IDENTITY CASCADE;

INSERT INTO users (full_name, email, password_hash, created_at)
VALUES
    (
        'Khang Test',
        'khang@gmail.com',
        '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm',
        NOW()
    ),
    (
        'Alice Nguyen',
        'alice@example.com',
        '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm',
        NOW()
    ),
    (
        'Bob Tran',
        'bob@example.com',
        '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm',
        NOW()
    );

INSERT INTO projects (name, description, owner_id, created_at)
SELECT
    'Task API Backend',
    'Sample project for auth and CRUD testing',
    u.id,
    NOW()
FROM users u
WHERE u.email = 'khang@gmail.com';

INSERT INTO projects (name, description, owner_id, created_at)
SELECT
    'Team Dashboard',
    'Second sample project for list and permission testing',
    u.id,
    NOW()
FROM users u
WHERE u.email = 'alice@example.com';

INSERT INTO tasks (project_id, title, description, status, assignee_id, created_at)
SELECT
    p.id,
    'Implement JWT login',
    'Login flow using bcrypt and JWT access token',
    'TODO',
    u.id,
    NOW()
FROM projects p
JOIN users u ON u.email = 'khang@gmail.com'
WHERE p.name = 'Task API Backend'
  AND p.owner_id = u.id;

INSERT INTO tasks (project_id, title, description, status, assignee_id, created_at)
SELECT
    p.id,
    'Create protected task endpoints',
    'Use middleware current_user context instead of manual user ID',
    'IN_PROGRESS',
    u.id,
    NOW()
FROM projects p
JOIN users u ON u.email = 'khang@gmail.com'
WHERE p.name = 'Task API Backend'
  AND p.owner_id = u.id;

INSERT INTO tasks (project_id, title, description, status, assignee_id, created_at)
SELECT
    p.id,
    'Prepare dashboard mock data',
    'Seed data for frontend preview and API smoke testing',
    'TODO',
    u.id,
    NOW()
FROM projects p
JOIN users u ON u.email = 'alice@example.com'
WHERE p.name = 'Team Dashboard'
  AND p.owner_id = u.id;

INSERT INTO tasks (project_id, title, description, status, assignee_id, created_at)
SELECT
    p.id,
    'Fix task delete flow',
    'Dedicated record for testing delete endpoints',
    'TODO',
    u.id,
    NOW()
FROM projects p
JOIN users u ON u.email = 'bob@example.com'
WHERE p.name = 'Team Dashboard'
  AND p.owner_id = u.id;
