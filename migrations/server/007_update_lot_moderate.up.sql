ALTER TABLE lots ADD COLUMN moderation_status text NOT NULL DEFAULT 'pending'
    CHECK (moderation_status IN ('pending', 'approved', 'rejected'));
ALTER TABLE lots ADD COLUMN rejection_reason text;