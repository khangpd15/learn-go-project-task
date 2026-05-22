BEGIN;

DELETE FROM tasks
WHERE title IN (
  'Design auth contract',
  'Implement JWT refresh flow',
  'Add audit logging middleware',
  'Build workload summary endpoint',
  'Add pagination + sorting',
  'Snapshot demo metrics',
  'Implement delta sync endpoint',
  'Retry policy for flaky clients',
  'Add sync contract tests',
  'Create smoke test pack',
  'Automate postman collection run',
  'Track flaky test cases',
  'Role matrix review',
  'Bulk user import',
  'Admin activity timeline',
  'Purge orphan task candidates',
  'Backfill missing created_at',
  'Data quality weekly report'
);

DELETE FROM projects
WHERE name IN (
  'Task API Backend',
  'Team Dashboard',
  'Mobile Sync Service',
  'QA Automation Hub',
  'Internal Admin Portal',
  'Data Cleanup Pipeline'
);

DELETE FROM users
WHERE email IN (
  'khang@gmail.com',
  'alice.nguyen@example.com',
  'bob.tran@example.com',
  'linh.pham@example.com',
  'minh.le@example.com',
  'hoa.do@example.com'
);

COMMIT;