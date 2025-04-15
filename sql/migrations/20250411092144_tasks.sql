-- +goose Up

CREATE TABLE collections(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL, 
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id)
);

CREATE TABLE cards(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    front TEXT NOT NULL,
    back TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    collection_id UUID NOT NULL REFERENCES collections(id)
);



-- +goose Down
DROP TABLE cards;
DROP TABLE collections;
