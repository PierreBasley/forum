ALTER TABLE likes
ADD COLUMN reaction TEXT NOT NULL DEFAULT 'like'
CHECK (reaction IN ('like', 'dislike'));
