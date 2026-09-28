-- create "organization_ssh_keys" table
CREATE TABLE "organization_ssh_keys" (
  "organization_id" uuid NOT NULL,
  "public_key" text NOT NULL,
  "encrypted_private_key" bytea NOT NULL,
  "nonce" bytea NOT NULL,
  "key_algorithm" text NOT NULL,
  "encryption_version" text NOT NULL,
  "fingerprint" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("organization_id"),
  CONSTRAINT "fk_organization_ssh_keys_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
