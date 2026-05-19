-- 000002_harden_constraints_indexes.down.sql

DROP INDEX IF EXISTS ux_users_email_lower;

DROP INDEX IF EXISTS idx_tasks_project_status;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_assignee_id;
DROP INDEX IF EXISTS idx_tasks_project_id;
DROP INDEX IF EXISTS idx_projects_owner_id;

ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_status;

-- Revert strict NOT NULL additions from v000002.
ALTER TABLE tasks ALTER COLUMN status DROP NOT NULL;
ALTER TABLE tasks ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE projects ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE users ALTER COLUMN created_at DROP NOT NULL;
