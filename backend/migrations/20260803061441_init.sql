-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" citext NULL,
  "password_hash" text NULL,
  "display_name" character varying(50) NOT NULL,
  "handle" character varying(30) NOT NULL,
  "avatar_asset_id" uuid NULL,
  "preferred_locale" character varying(10) NOT NULL DEFAULT 'ja',
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "level" integer NOT NULL DEFAULT 1,
  "experience_points" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "anonymized_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "chk_users_experience_points" CHECK (experience_points >= 0),
  CONSTRAINT "chk_users_level" CHECK (level >= 1),
  CONSTRAINT "chk_users_status" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'suspended'::character varying, 'deleted'::character varying])::text[]))
);
-- Create "blocks" table
CREATE TABLE "blocks" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "blocker_user_id" uuid NOT NULL,
  "blocked_user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_blocks_blocked" FOREIGN KEY ("blocked_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_blocks_blocker" FOREIGN KEY ("blocker_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_blocks_blocked_user_id" to table: "blocks"
CREATE INDEX "idx_blocks_blocked_user_id" ON "blocks" ("blocked_user_id");
-- Create index "ux_blocks_pair" to table: "blocks"
CREATE UNIQUE INDEX "ux_blocks_pair" ON "blocks" ("blocker_user_id", "blocked_user_id");
-- Create "friendships" table
CREATE TABLE "friendships" (
  "user_low_id" uuid NOT NULL,
  "user_high_id" uuid NOT NULL,
  "requested_by_user_id" uuid NOT NULL,
  "status" character varying(20) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("user_low_id", "user_high_id"),
  CONSTRAINT "fk_friendships_user_high" FOREIGN KEY ("user_high_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_friendships_user_low" FOREIGN KEY ("user_low_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_friendships_status" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'accepted'::character varying, 'rejected'::character varying])::text[]))
);
-- Create index "idx_friendships_high_status" to table: "friendships"
CREATE INDEX "idx_friendships_high_status" ON "friendships" ("user_high_id", "status");
-- Create "media_assets" table
CREATE TABLE "media_assets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "owner_user_id" uuid NOT NULL,
  "purpose" character varying(30) NOT NULL,
  "storage_key" text NOT NULL,
  "original_filename" character varying(255) NOT NULL,
  "mime_type" character varying(100) NOT NULL,
  "size_bytes" bigint NOT NULL,
  "width" integer NULL,
  "height" integer NULL,
  "checksum_sha256" character(64) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "created_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_media_assets_owner" FOREIGN KEY ("owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_media_assets_purpose" CHECK ((purpose)::text = 'avatar'::text),
  CONSTRAINT "chk_media_assets_size_bytes" CHECK (size_bytes > 0),
  CONSTRAINT "chk_media_assets_status" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'deleted'::character varying])::text[]))
);
-- Create index "idx_media_assets_owner" to table: "media_assets"
CREATE INDEX "idx_media_assets_owner" ON "media_assets" ("owner_user_id", "purpose", "status");
-- Create index "idx_media_assets_storage_key" to table: "media_assets"
CREATE UNIQUE INDEX "idx_media_assets_storage_key" ON "media_assets" ("storage_key");
-- Create "oauth_accounts" table
CREATE TABLE "oauth_accounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "provider" character varying(30) NOT NULL,
  "provider_account_id" character varying(255) NOT NULL,
  "provider_email" citext NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_oauth_accounts_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_oauth_accounts_provider" CHECK ((provider)::text = ANY ((ARRAY['google'::character varying, 'github'::character varying, '42'::character varying])::text[]))
);
-- Create index "idx_oauth_accounts_user_id" to table: "oauth_accounts"
CREATE INDEX "idx_oauth_accounts_user_id" ON "oauth_accounts" ("user_id");
-- Create index "ux_oauth_provider_account" to table: "oauth_accounts"
CREATE UNIQUE INDEX "ux_oauth_provider_account" ON "oauth_accounts" ("provider", "provider_account_id");
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "token_hash" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "last_seen_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_sessions_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sessions_token_hash" to table: "sessions"
CREATE UNIQUE INDEX "idx_sessions_token_hash" ON "sessions" ("token_hash");
-- Create index "idx_sessions_user_expires" to table: "sessions"
CREATE INDEX "idx_sessions_user_expires" ON "sessions" ("user_id", "expires_at");
