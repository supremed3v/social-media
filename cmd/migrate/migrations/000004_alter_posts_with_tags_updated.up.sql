ALTER TABLE posts
ADD COLUMN tags VARCHAR(100) [];
ALTER TABLE posts ADD COLUMN updatedAt timestamp(0) with time zone NOT NULL DEFAULT NOW()