-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE "google_accounts" ADD COLUMN token_type TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE "google_accounts" DROP COLUMN token_type;
-- +goose StatementEnd
