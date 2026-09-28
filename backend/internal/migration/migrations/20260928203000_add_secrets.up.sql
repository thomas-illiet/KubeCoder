CREATE TABLE "secrets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "scope" text NOT NULL,
  "organization_id" uuid NULL,
  "variable_name" text NOT NULL,
  "description" text NOT NULL,
  "encrypted_value" bytea NOT NULL,
  "nonce" bytea NOT NULL,
  "fingerprint" text NOT NULL,
  "encryption_version" text NOT NULL,
  "expires_at" timestamptz NULL,
  "value_replaced_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "secrets_scope_owner_check" CHECK (
    (scope = 'PLATFORM' AND organization_id IS NULL) OR
    (scope = 'ORGANIZATION' AND organization_id IS NOT NULL)
  ),
  CONSTRAINT "secrets_organization_fk" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX "secrets_platform_variable_name_key" ON "secrets" ("variable_name") WHERE scope = 'PLATFORM';
CREATE UNIQUE INDEX "secrets_organization_variable_name_key" ON "secrets" ("organization_id", "variable_name") WHERE scope = 'ORGANIZATION';
CREATE INDEX "idx_secrets_organization_id" ON "secrets" ("organization_id");

CREATE FUNCTION "enforce_secret_platform_precedence"() RETURNS trigger AS $$
BEGIN
  PERFORM pg_advisory_xact_lock(hashtextextended(NEW.variable_name, 0));
  IF NEW.scope = 'PLATFORM' AND EXISTS (
    SELECT 1 FROM secrets WHERE scope = 'ORGANIZATION' AND variable_name = NEW.variable_name AND id <> NEW.id
  ) THEN
    RAISE EXCEPTION 'secret variable conflicts with an organization secret' USING ERRCODE = '23505';
  END IF;
  IF NEW.scope = 'ORGANIZATION' AND EXISTS (
    SELECT 1 FROM secrets WHERE scope = 'PLATFORM' AND variable_name = NEW.variable_name AND id <> NEW.id
  ) THEN
    RAISE EXCEPTION 'secret variable conflicts with a platform secret' USING ERRCODE = '23505';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER "secrets_platform_precedence_check"
BEFORE INSERT OR UPDATE OF scope, variable_name, organization_id ON "secrets"
FOR EACH ROW EXECUTE FUNCTION "enforce_secret_platform_precedence"();

CREATE TABLE "secret_bindings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "secret_id" uuid NOT NULL,
  "target_type" text NOT NULL,
  "agent_id" uuid NULL,
  "repository_id" uuid NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "secret_bindings_target_check" CHECK (
    (target_type = 'AGENT' AND agent_id IS NOT NULL AND repository_id IS NULL) OR
    (target_type = 'REPOSITORY' AND agent_id IS NULL AND repository_id IS NOT NULL)
  ),
  CONSTRAINT "secret_bindings_secret_fk" FOREIGN KEY ("secret_id") REFERENCES "secrets" ("id") ON DELETE CASCADE,
  CONSTRAINT "secret_bindings_agent_fk" FOREIGN KEY ("agent_id") REFERENCES "agents" ("id") ON DELETE CASCADE,
  CONSTRAINT "secret_bindings_repository_fk" FOREIGN KEY ("repository_id") REFERENCES "repositories" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX "secret_bindings_agent_key" ON "secret_bindings" ("secret_id", "agent_id") WHERE agent_id IS NOT NULL;
CREATE UNIQUE INDEX "secret_bindings_repository_key" ON "secret_bindings" ("secret_id", "repository_id") WHERE repository_id IS NOT NULL;
CREATE INDEX "idx_secret_bindings_secret_id" ON "secret_bindings" ("secret_id");
CREATE INDEX "idx_secret_bindings_agent_id" ON "secret_bindings" ("agent_id");
CREATE INDEX "idx_secret_bindings_repository_id" ON "secret_bindings" ("repository_id");
