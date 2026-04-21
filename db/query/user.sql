-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3 ,$4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1
LIMIT 1;

-- name: UpdateUser :one
-- UpdateUser 支持“部分更新”：
-- 1) sqlc.narg(field) 会生成可空参数（sql.NullString）。
-- 2) COALESCE(a, b) 的意思是：a 不为 NULL 就用 a，否则用 b。
-- 3) 组合起来就是：
--    - 传了新值 -> 更新该字段
--    - 没传值(NULL) -> 保留数据库原值
UPDATE users
SET 
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
  password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
  full_name = COALESCE(sqlc.narg(full_name), full_name),
  email = COALESCE(sqlc.narg(email), email)
WHERE
  username = sqlc.arg(username)
RETURNING *;
