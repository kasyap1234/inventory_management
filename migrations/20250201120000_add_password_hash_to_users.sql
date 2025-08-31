-- Add password_hash column to users table for password-based authentication
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

-- Add index on email for faster authentication
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Add comment explaining the column purpose
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password for user authentication';