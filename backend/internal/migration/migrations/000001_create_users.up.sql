CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer text NOT NULL,
    subject text NOT NULL,
    username text NOT NULL,
    display_name text NOT NULL,
    email text NOT NULL,
    is_admin boolean NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX users_issuer_subject_key ON users (issuer, subject);
