CREATE TABLE devices (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    type VARCHAR(32) NOT NULL,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE  NOT NULL,
    client_uuid uuid NOT NULL,
    platform VARCHAR(32) NOT NULL,
    platform_version VARCHAR(16) NOT NULL,
    state_version uuid NOT NULL,
    app_version VARCHAR(16) NOT NULL,
    last_active_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX idx_uniq_client_uuid ON devices(client_uuid, user_id);
