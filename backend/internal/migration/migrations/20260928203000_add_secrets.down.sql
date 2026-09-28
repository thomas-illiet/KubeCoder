DROP TABLE "secret_bindings";
DROP TABLE "secrets";
DROP FUNCTION "enforce_secret_platform_precedence"();
ALTER TABLE "repositories" DROP COLUMN "secret_mode";
