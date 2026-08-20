-- Create "matches" table
CREATE TABLE "matches" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "mode" character varying(20) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'in_progress',
  "result_type" character varying(20) NULL,
  "total_moves" integer NOT NULL DEFAULT 0,
  "started_at" timestamptz NOT NULL,
  "finished_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "chk_matches_mode" CHECK ((mode)::text = ANY ((ARRAY['ranked'::character varying, 'casual'::character varying, 'ai'::character varying, 'friend'::character varying])::text[])),
  CONSTRAINT "chk_matches_result_type" CHECK ((result_type IS NULL) OR ((result_type)::text = ANY ((ARRAY['goal'::character varying, 'resign'::character varying, 'timeout'::character varying, 'draw'::character varying, 'abort'::character varying])::text[]))),
  CONSTRAINT "chk_matches_status" CHECK ((status)::text = ANY ((ARRAY['in_progress'::character varying, 'finished'::character varying, 'aborted'::character varying])::text[])),
  CONSTRAINT "chk_matches_total_moves" CHECK (total_moves >= 0)
);
-- Create index "idx_matches_status_finished" to table: "matches"
CREATE INDEX "idx_matches_status_finished" ON "matches" ("status", "finished_at" DESC);
-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "rating" integer NOT NULL DEFAULT 1200;
-- Create index "idx_users_rating" to table: "users"
CREATE INDEX "idx_users_rating" ON "users" ("rating" DESC);
-- Create "match_participants" table
CREATE TABLE "match_participants" (
  "match_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "seat" smallint NOT NULL,
  "outcome" character varying(10) NOT NULL,
  "rating_before" integer NOT NULL,
  "rating_after" integer NOT NULL,
  "xp_gained" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("match_id", "user_id"),
  CONSTRAINT "fk_match_participants_match" FOREIGN KEY ("match_id") REFERENCES "matches" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_match_participants_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_match_participants_outcome" CHECK ((outcome)::text = ANY ((ARRAY['win'::character varying, 'loss'::character varying, 'draw'::character varying])::text[])),
  CONSTRAINT "chk_match_participants_seat" CHECK (seat = ANY (ARRAY[0, 1])),
  CONSTRAINT "chk_match_participants_xp_gained" CHECK (xp_gained >= 0)
);
-- Create index "idx_match_participants_user" to table: "match_participants"
CREATE INDEX "idx_match_participants_user" ON "match_participants" ("user_id");
-- Create "user_achievements" table
CREATE TABLE "user_achievements" (
  "user_id" uuid NOT NULL,
  "code" character varying(50) NOT NULL,
  "unlocked_at" timestamptz NOT NULL,
  PRIMARY KEY ("user_id", "code"),
  CONSTRAINT "fk_user_achievements_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- ここから下は GORM のタグで表現できないぶんの手書き。
-- 複数列にまたがる CHECK は gorm の check タグでは 1 列にしか付けられないため。

-- 決着済みの対戦は結果と終了時刻を必ず持つ。
-- 進行中の行と完了済みの行が同じテーブルに混ざるので、
-- 「finished なのに result_type が NULL」を DB 側で弾く。
ALTER TABLE "matches" ADD CONSTRAINT "chk_matches_finished_has_result"
  CHECK ("status" <> 'finished' OR ("result_type" IS NOT NULL AND "finished_at" IS NOT NULL));

-- 対戦の終了時刻は開始時刻より前にならない。
ALTER TABLE "matches" ADD CONSTRAINT "chk_matches_finished_after_started"
  CHECK ("finished_at" IS NULL OR "finished_at" >= "started_at");

-- レーティングは負にならない。
ALTER TABLE "match_participants" ADD CONSTRAINT "chk_match_participants_rating"
  CHECK ("rating_before" >= 0 AND "rating_after" >= 0);
