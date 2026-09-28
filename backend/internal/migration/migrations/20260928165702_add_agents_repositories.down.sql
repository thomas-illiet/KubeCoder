-- reverse: create index "idx_repository_agent_bindings_agent_id" to table: "repository_agent_bindings"
DROP INDEX "idx_repository_agent_bindings_agent_id";
-- reverse: create "repository_agent_bindings" table
DROP TABLE "repository_agent_bindings";
-- reverse: create index "idx_agents_name" to table: "agents"
DROP INDEX "idx_agents_name";
-- reverse: create "agents" table
DROP TABLE "agents";
-- reverse: create index "repositories_organization_clone_url_key" to table: "repositories"
DROP INDEX "repositories_organization_clone_url_key";
-- reverse: create "repositories" table
DROP TABLE "repositories";
