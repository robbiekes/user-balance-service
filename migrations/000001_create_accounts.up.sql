CREATE TABLE IF NOT EXISTS users
(
    id SERIAL NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts
(
    id SERIAL NOT NULL UNIQUE,
    balance INT
);

CREATE TABLE IF NOT EXISTS history
(
    id SERIAL NOT NULL UNIQUE PRIMARY KEY,
    type VARCHAR(255),
    description VARCHAR(255),
    amount INT,
    account_id INT NOT NULL
        REFERENCES accounts (id) ON DELETE RESTRICT,
    date TIMESTAMP NOT NULL
);