-- create "organizations" table
CREATE TABLE "organizations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_organizations_slug" to table: "organizations"
CREATE UNIQUE INDEX "idx_organizations_slug" ON "organizations" ("slug");
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "preferred_organization_id" uuid NULL, ADD CONSTRAINT "fk_users_preferred_organization" FOREIGN KEY ("preferred_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "idx_users_preferred_organization_id" to table: "users"
CREATE INDEX "idx_users_preferred_organization_id" ON "users" ("preferred_organization_id");
-- create "organization_memberships" table
CREATE TABLE "organization_memberships" (
  "organization_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("organization_id", "user_id"),
  CONSTRAINT "fk_organization_memberships_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_organization_memberships_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
