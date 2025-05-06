-- name: CreatePlayerSalaries :one
INSERT INTO nba_player_salaries (
  name, salary2025, salary2026, salary2027, salary2028, salary2029
) VALUES (
$1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: CreateTeamSalaries :one
INSERT INTO nba_team_salaries (
  name, salary2025, salary2026, salary2027, salary2028, salary2029
) VALUES (
$1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetPlayersSalaries :many
SELECT * FROM nba_player_salaries
ORDER BY name;

-- name: GetTeamsSalaries :many
SELECT * FROM nba_team_salaries
ORDER BY name;

-- name: GetPlayersSalaryByID :one
SELECT * FROM nba_player_salaries
WHERE id = $1 LIMIT 1;

-- name: GetTeamSalaryByID :one
SELECT * FROM nba_team_salaries
WHERE id = $1 LIMIT 1;


