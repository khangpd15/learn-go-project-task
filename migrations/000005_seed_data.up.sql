BEGIN;


INSERT INTO users (full_name, email, password_hash, created_at)
VALUES
('Nguyen Khang', 'khang@gmail.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Alice Nguyen', 'alice.nguyen@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Bob Tran', 'bob.tran@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Linh Pham', 'linh.pham@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Minh Le', 'minh.le@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW()),
('Hoa Do', 'hoa.do@example.com', '$2a$10$3v.Wye1FxZm53uNz9rpVUOIu.wOovLaZIhNyNg8vU1Ula6.6fcCZm', NOW())
ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (name, description, owner_id, created_at)
SELECT
    v.name,
    v.description,
    u.id,
    NOW()
FROM (
    VALUES
        ('Task API Backend', 'Core backend API for authentication and task management', 'khang@gmail.com'),
        ('Team Dashboard', 'Dashboard API for team workload and progress', 'alice.nguyen@example.com'),
        ('Mobile Sync Service', 'Sync endpoint layer for mobile clients', 'bob.tran@example.com'),
        ('QA Automation Hub', 'Automation support service for smoke and regression tests', 'linh.pham@example.com'),
        ('Internal Admin Portal', 'Administrative APIs for user and project governance', 'minh.le@example.com'),
        ('Data Cleanup Pipeline', 'Background data maintenance and consistency checks', 'hoa.do@example.com')
) AS v(name, description, owner_email)
JOIN users u ON u.email = v.owner_email;

-- -----------------------------------------------------------------------------
-- 3) TASKS
-- Notes:
-- - project_id is resolved by (project name + owner) to avoid ambiguous matches.
-- - assignee_id is nullable by schema; we intentionally keep some NULL assignees.
-- - status only uses valid values: TODO, IN_PROGRESS, DONE.
-- - created_at uses NOW().
-- -----------------------------------------------------------------------------
INSERT INTO tasks (project_id, title, description, status, assignee_id, created_at)
SELECT
    p.id,
    t.title,
    t.description,
    t.status,
    au.id,
    NOW()
FROM (
    VALUES
        ('Task API Backend',      'khang@gmail.com',              'Design auth contract',             'Finalize login/register request and response schema',                          'DONE',        'khang@gmail.com'),
        ('Task API Backend',      'khang@gmail.com',              'Implement JWT refresh flow',       'Add refresh token endpoint and token rotation strategy',                      'IN_PROGRESS', 'alice.nguyen@example.com'),
        ('Task API Backend',      'khang@gmail.com',              'Add audit logging middleware',     'Track request-id, actor, and endpoint in structured logs',                    'TODO',        NULL),

        ('Team Dashboard',        'alice.nguyen@example.com',     'Build workload summary endpoint',  'Aggregate TODO/IN_PROGRESS/DONE tasks by assignee',                           'IN_PROGRESS', 'bob.tran@example.com'),
        ('Team Dashboard',        'alice.nguyen@example.com',     'Add pagination + sorting',         'Support stable list ordering and paging metadata',                            'TODO',        'linh.pham@example.com'),
        ('Team Dashboard',        'alice.nguyen@example.com',     'Snapshot demo metrics',            'Prepare deterministic numbers for product demo',                              'DONE',        'alice.nguyen@example.com'),

        ('Mobile Sync Service',   'bob.tran@example.com',         'Implement delta sync endpoint',    'Return only changes since provided sync timestamp',                           'IN_PROGRESS', 'minh.le@example.com'),
        ('Mobile Sync Service',   'bob.tran@example.com',         'Retry policy for flaky clients',   'Introduce idempotent retries and exponential backoff docs',                   'TODO',        'hoa.do@example.com'),
        ('Mobile Sync Service',   'bob.tran@example.com',         'Add sync contract tests',          'Cover clock drift and duplicated update events',                              'DONE',        'bob.tran@example.com'),

        ('QA Automation Hub',     'linh.pham@example.com',        'Create smoke test pack',           'Core endpoint checks for auth, projects, and tasks',                          'DONE',        'linh.pham@example.com'),
        ('QA Automation Hub',     'linh.pham@example.com',        'Automate postman collection run',  'Run collection in CI with environment-specific variables',                     'IN_PROGRESS', 'khang@gmail.com'),
        ('QA Automation Hub',     'linh.pham@example.com',        'Track flaky test cases',           'Persist intermittent failures for triage dashboard',                          'TODO',        NULL),

        ('Internal Admin Portal', 'minh.le@example.com',          'Role matrix review',               'Review allowed actions for admin, member, and viewer roles',                  'TODO',        'alice.nguyen@example.com'),
        ('Internal Admin Portal', 'minh.le@example.com',          'Bulk user import',                 'Add CSV import endpoint with validation report',                              'IN_PROGRESS', 'minh.le@example.com'),
        ('Internal Admin Portal', 'minh.le@example.com',          'Admin activity timeline',          'Expose activity events with date range filter',                               'DONE',        'hoa.do@example.com'),

        ('Data Cleanup Pipeline', 'hoa.do@example.com',           'Purge orphan task candidates',     'Identify and remove logically invalid data snapshots',                         'TODO',        'bob.tran@example.com'),
        ('Data Cleanup Pipeline', 'hoa.do@example.com',           'Backfill missing created_at',      'Fix legacy rows where created_at was accidentally null before hardening',     'DONE',        'hoa.do@example.com'),
        ('Data Cleanup Pipeline', 'hoa.do@example.com',           'Data quality weekly report',       'Generate project/task consistency report for engineering lead',                'IN_PROGRESS', NULL)
) AS t(project_name, owner_email, title, description, status, assignee_email)
JOIN users ou ON ou.email = t.owner_email
JOIN projects p ON p.name = t.project_name AND p.owner_id = ou.id
LEFT JOIN users au ON au.email = t.assignee_email;

COMMIT;