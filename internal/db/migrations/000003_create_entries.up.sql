CREATE TABLE entries (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    type VARCHAR(32) NOT NULL,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    device_id uuid REFERENCES devices(id) ON DELETE CASCADE NOT NULL,
    title VARCHAR(255) NOT NULL
)