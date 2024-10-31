-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, is_chirpy_red, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    false,
    $1,
    $2
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users where email = $1;

-- name: UpdateUser :one
UPDATE users SET
    email = $1,
    hashed_password = $2,
    updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: EnableChirpyRed :exec
UPDATE users SET
    is_chirpy_red = true,
    updated_at = NOW()
WHERE id = $1;
