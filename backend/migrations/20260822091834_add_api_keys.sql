-- Create "api_keys" table
CREATE TABLE "api_keys" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "name" character varying(50) NOT NULL,
  "key_prefix" character varying(16) NOT NULL,
  "key_hash" character(64) NOT NULL,
  "scopes" character varying(100) NOT NULL,
  "expires_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "last_used_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_api_keys_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_api_keys_key_hash" to table: "api_keys"
CREATE UNIQUE INDEX "idx_api_keys_key_hash" ON "api_keys" ("key_hash");
-- Create index "idx_api_keys_user_id" to table: "api_keys"
CREATE INDEX "idx_api_keys_user_id" ON "api_keys" ("user_id");
