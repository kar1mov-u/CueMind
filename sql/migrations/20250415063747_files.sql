-- +goose Up
CREATE TABLE files(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id),
    user_id UUID NOT NULL REFERENCES users(id), 
    file_name TEXT,
    format TEXT, 
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE files;