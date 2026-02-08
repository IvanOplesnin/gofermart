-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    "login" VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    CONSTRAINT users_login_uk UNIQUE ("login")
);

CREATE TABLE IF NOT EXISTS order_numbers (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    "number" VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    "status" VARCHAR(30) NOT NULL,
    accrual NUMERIC(12, 2),
    uploaded_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT order_status_chk CHECK("status" IN ('PROCESSING', 'PROCESSED', 'INVALID', 'NEW')),

    CONSTRAINT order_numbers_number_uk UNIQUE ("number"),
    CONSTRAINT order_numbers_user_fk FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS order_numbers_user_id_uploaded_at_idx ON order_numbers(user_id, uploaded_at);


CREATE TABLE IF NOT EXISTS user_balance (
        id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        user_id INTEGER NOT NULL,
        balance NUMERIC(12,2) NOT NULL DEFAULT 0,
        withdrawn NUMERIC(12,2) NOT NULL DEFAULT 0,

        CHECK (balance >= 0 AND withdrawn >= 0),
        CONSTRAINT user_balance_user_uk UNIQUE (user_id),
        CONSTRAINT user_balance_user_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS withdraws (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL,
    order_number VARCHAR(255) NOT NULL,
    summa NUMERIC(12,2) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL,


    CONSTRAINT withdraws_user_fk FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT withdraws_order_number_uk UNIQUE (order_number)
);
CREATE INDEX IF NOT EXISTS withdraws_user_id_processed_at_idx ON withdraws(user_id, processed_at);


-- +goose StatementEnd  


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdraws;
DROP TABLE IF EXISTS order_numbers;
DROP TABLE IF EXISTS user_balance;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
