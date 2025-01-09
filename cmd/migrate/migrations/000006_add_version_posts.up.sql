DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'posts' 
        AND column_name = 'version'
    ) THEN
        ALTER TABLE posts ADD COLUMN version INT NOT NULL DEFAULT 0;
    END IF;
END $$;
