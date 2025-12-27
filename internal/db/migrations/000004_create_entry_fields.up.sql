CREATE TABLE entry_fields (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    entry_id uuid REFERENCES entries(id) ON DELETE CASCADE NOT NULL,
    key VARCHAR(255) NOT NULL,
    value bytea,
    is_private bool DEFAULT false NOT NULL
);

CREATE INDEX idx_private ON entry_fields(entry_id, is_private);