-- reverse: create "organization_memberships" table
DROP TABLE "organization_memberships";
-- reverse: create index "idx_users_preferred_organization_id" to table: "users"
DROP INDEX "idx_users_preferred_organization_id";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "fk_users_preferred_organization", DROP COLUMN "preferred_organization_id";
-- reverse: create index "idx_organizations_slug" to table: "organizations"
DROP INDEX "idx_organizations_slug";
-- reverse: create "organizations" table
DROP TABLE "organizations";
