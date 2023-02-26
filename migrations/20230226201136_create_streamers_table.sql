-- +goose Up
-- +goose StatementBegin
CREATE TABLE streamers (
  id SERIAL PRIMARY KEY,
  wallet_address VARCHAR(90) UNIQUE NOT NULL,
  client_id VARCHAR(90),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX client_id_index ON streamers (client_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE streamers;
-- +goose StatementEnd
