-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id             UUID PRIMARY KEY,
    last_name      VARCHAR(255) NOT NULL,
    first_name     VARCHAR(255) NOT NULL,
    middle_name    VARCHAR(255),
    email          VARCHAR(255) NOT NULL,
    email_confirmed BOOLEAN     NOT NULL DEFAULT FALSE,
    password_hash  VARCHAR(255) NOT NULL,
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMPTZ
);

CREATE UNIQUE INDEX users_email_active_uniq ON users(email) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
