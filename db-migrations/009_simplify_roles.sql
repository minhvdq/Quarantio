-- Drop the old role constraint and replace with owner/user only.
ALTER TABLE org_members DROP CONSTRAINT IF EXISTS org_members_role_check;
ALTER TABLE org_members ADD CONSTRAINT org_members_role_check
    CHECK (role IN ('owner', 'user'));

-- Demote any existing manager/monitor rows to user before constraint applies.
UPDATE org_members SET role = 'user' WHERE role IN ('manager', 'monitor');

-- Drop release_requests role constraints that referenced monitor/manager.
-- The table stays; only owners action HIGH review requests now.
