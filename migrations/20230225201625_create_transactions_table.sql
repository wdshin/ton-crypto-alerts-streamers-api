-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  tx_hash VARCHAR(64) UNIQUE NOT NULL,
  sign VARCHAR(64) NOT NULL,
  wallet_address VARCHAR(90),
  amount BIGINT NOT NULL,
  lt BIGINT NOT NULL,
  acked BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX sign_index ON transactions (sign);
CREATE INDEX tx_hash_index ON transactions (tx_hash);
CREATE INDEX lt_index ON transactions (lt);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX tx_hash_index;
DROP INDEX lt_index;
DROP INDEX sign_index;

DROP TABLE transactions;
-- +goose StatementEnd
