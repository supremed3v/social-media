DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE constraint_type = 'FOREIGN KEY'
        AND table_name = 'posts'
        AND constraint_name = 'fk_user'
    ) THEN
        ALTER TABLE posts
        ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id);
    END IF;
END $$;
