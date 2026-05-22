-- 000004_reset_and_seed_sample_data.down.sql
-- Remove the sample seed data inserted by 000004.

TRUNCATE TABLE tasks, projects, users RESTART IDENTITY CASCADE;
