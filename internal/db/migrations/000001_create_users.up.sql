CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    username VARCHAR(32) NOT NULL,
    encrypted_password VARCHAR(60) NOT NULL
);

CREATE UNIQUE INDEX idx_uniq_username ON users(username);
