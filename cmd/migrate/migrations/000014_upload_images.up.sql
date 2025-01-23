CREATE TABLE IF NOT EXISTS images(
    id bigserial PRIMARY KEY,
    url text NOT NULL,
    createdAt timestamp(0) with time zone NOT NULL DEFAULT NOW()
);