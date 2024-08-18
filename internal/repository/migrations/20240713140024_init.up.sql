CREATE TABLE todos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  description TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT Now(),
  updated_at TIMESTAMP,
  completed_at TIMESTAMP,
  deleted_at TIMESTAMP
);
