DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'posts' 
        AND column_name = 'tags'
    ) THEN
        ALTER TABLE posts ADD COLUMN tags VARCHAR(100);
    END IF;

    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'posts' 
        AND column_name = 'updatedat'
    ) THEN
        ALTER TABLE posts ADD COLUMN updatedAt TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW();
    END IF;
END $$;
