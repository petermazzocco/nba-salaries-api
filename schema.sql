CREATE TABLE nba_player_salaries (
    id              BIGSERIAL PRIMARY KEY,
    name            text      NOT NULL,
    salary2025      text,
    salary2026      text,
    salary2027      text,
    salary2028      text,
    salary2029      text
);

CREATE TABLE nba_team_salaries (
    id              BIGSERIAL PRIMARY KEY,
    name            text      NOT NULL,
    salary2025      text,
    salary2026      text,
    salary2027      text,
    salary2028      text,
    salary2029      text
);
