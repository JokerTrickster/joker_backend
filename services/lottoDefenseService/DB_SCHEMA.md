# Lotto Defense Service – Database Schema

Game data lives in the same MySQL database as other services. Migrations are in repo root `migrations/`.

## Tables

### game_rounds

| Column      | Type         | Description                    |
|------------|--------------|--------------------------------|
| id         | BIGINT PK    | Auto-increment                 |
| user_id    | BIGINT       | FK to users(id)               |
| status     | VARCHAR(20) | `active` or `completed`       |
| score      | INT UNSIGNED | Null until round completed    |
| started_at | TIMESTAMP    | When round started            |
| ended_at   | TIMESTAMP    | When round ended              |
| created_at | TIMESTAMP    |                                |
| updated_at | TIMESTAMP    |                                |

### lotto_draws

| Column    | Type      | Description              |
|-----------|-----------|--------------------------|
| id        | BIGINT PK | Auto-increment           |
| round_id  | BIGINT    | FK to game_rounds(id), unique |
| numbers   | JSON      | Array of 6 integers (1–45) |
| created_at| TIMESTAMP |                          |

One draw per round, created when the round is completed.
