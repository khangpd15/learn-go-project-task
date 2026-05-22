INSERT INTO users (full_name, email, password_hash, created_at)
VALUES (
  'Khang Test',
  'khang@gmail.com',
  '$2a$10$7qW/.r7GKEUGbl7D5YLtK.u9Gx1U.FCKLP.g0L9NyxrwtKchqXsYm',
  NOW()
)
ON CONFLICT (email) DO NOTHING;