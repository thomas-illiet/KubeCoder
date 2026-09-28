-- create "repositories" table
CREATE TABLE "repositories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "organization_id" uuid NOT NULL,
  "name" text NOT NULL,
  "provider" text NOT NULL,
  "clone_url" text NOT NULL,
  "default_branch" text NOT NULL,
  "include_submodules" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_repositories_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "repositories_organization_clone_url_key" to table: "repositories"
CREATE UNIQUE INDEX "repositories_organization_clone_url_key" ON "repositories" ("organization_id", "clone_url");
-- create "agents" table
CREATE TABLE "agents" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NOT NULL,
  "runtime_adapter" text NOT NULL,
  "runtime_version" text NOT NULL,
  "image" text NOT NULL,
  "provider" text NOT NULL,
  "model" text NOT NULL,
  "system_prompt" text NOT NULL,
  "capabilities" jsonb NOT NULL DEFAULT '[]',
  "cpu_millis" bigint NOT NULL,
  "memory_mb" bigint NOT NULL,
  "storage_mb" bigint NOT NULL,
  "max_duration_secs" bigint NOT NULL,
  "active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_agents_name" to table: "agents"
CREATE UNIQUE INDEX "idx_agents_name" ON "agents" ("name");
-- create "repository_agent_bindings" table
CREATE TABLE "repository_agent_bindings" (
  "repository_id" uuid NOT NULL,
  "agent_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("repository_id"),
  CONSTRAINT "fk_repository_agent_bindings_agent" FOREIGN KEY ("agent_id") REFERENCES "agents" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "fk_repository_agent_bindings_repository" FOREIGN KEY ("repository_id") REFERENCES "repositories" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_repository_agent_bindings_agent_id" to table: "repository_agent_bindings"
CREATE INDEX "idx_repository_agent_bindings_agent_id" ON "repository_agent_bindings" ("agent_id");
