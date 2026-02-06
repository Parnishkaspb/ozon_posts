CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    surname text NOT NULL,
    login text NOT NULL UNIQUE,
    password text NOT NULL
);

INSERT INTO users (name, surname, login, password) VALUES ('Иван', 'Грозный', 'Ivan', 'MoscowNeverSleep')

CREATE TABLE posts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id uuid NOT NULL,
    text text NOT NULL,
    without_comment boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT posts_author_fk FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);