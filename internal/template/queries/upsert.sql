INSERT INTO templates (name, content, created_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
  content = excluded.content,
  updated_at = excluded.updated_at

