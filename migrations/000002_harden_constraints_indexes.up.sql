-- 000002_harden_constraints_indexes.up.sql
-- Backward-compatible hardening for constraints and indexes.

-- 1) Keep timestamp columns consistent: default already exists in v000001,
--    here we enforce NOT NULL after filling historical NULL data.
UPDATE users SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE projects SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE tasks SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;

ALTER TABLE users ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE projects ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE tasks ALTER COLUMN created_at SET NOT NULL;

-- 2) Normalize existing task status data before adding strong constraints.
UPDATE tasks
SET status = 'TODO'
WHERE status IS NULL
   OR status NOT IN ('TODO', 'IN_PROGRESS', 'DONE');

ALTER TABLE tasks ALTER COLUMN status SET DEFAULT 'TODO';
ALTER TABLE tasks ALTER COLUMN status SET NOT NULL;

ALTER TABLE tasks
ADD CONSTRAINT chk_tasks_status
CHECK (status IN ('TODO', 'IN_PROGRESS', 'DONE'));

-- 3) Add indexes for common FK/filter paths.
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks(project_id, status);

-- 4) Enforce case-insensitive uniqueness for email (still keep existing UNIQUE(email)).
--    This blocks duplicate logical emails like Test@Mail.com vs test@mail.com.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM users u1
        JOIN users u2
          ON lower(u1.email) = lower(u2.email)
         AND u1.id < u2.id
    ) THEN
        RAISE EXCEPTION 'Cannot add case-insensitive unique email index: duplicate emails exist when lower-cased';
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email_lower ON users (lower(email));
