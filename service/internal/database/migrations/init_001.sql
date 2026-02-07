CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    surname text NOT NULL,
    login text NOT NULL UNIQUE,
    password text NOT NULL
);

INSERT INTO users (name, surname, login, password) VALUES ('Иван', 'Грозный', 'Ivan', 'MoscowNeverSleep');

CREATE TABLE posts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id uuid NOT NULL,
    text text NOT NULL,
    without_comment boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT posts_author_fk FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE comments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id uuid NOT NULL,
    author_id uuid NOT NULL,
    parent_id uuid NULL,
    text text NOT NULL CHECK (char_length(text) <= 2000),
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT comments_author_fk FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT comments_post_fk FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT comments_parent_fk FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX comments_post_root_idx
    ON comments (post_id, created_at, id)
    WHERE parent_id IS NULL;

CREATE INDEX comments_parent_idx
    ON comments (parent_id, created_at, id);