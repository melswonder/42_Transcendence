-- Modify "match_participants" table
-- 対局開始時に participants を先に挿入できるよう、決着時にしか決まらない列を NULL 許容にする。
ALTER TABLE "match_participants" DROP CONSTRAINT "chk_match_participants_outcome", ADD CONSTRAINT "chk_match_participants_outcome" CHECK ((outcome IS NULL) OR ((outcome)::text = ANY ((ARRAY['win'::character varying, 'loss'::character varying, 'draw'::character varying])::text[]))), ALTER COLUMN "outcome" DROP NOT NULL, ALTER COLUMN "rating_after" DROP NOT NULL;
-- Create "match_actions" table
CREATE TABLE "match_actions" (
  "match_id" uuid NOT NULL,
  "action_seq" integer NOT NULL,
  "action_id" uuid NOT NULL,
  "actor_seat" smallint NOT NULL,
  "action_type" character varying(20) NOT NULL,
  "payload" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("match_id", "action_seq"),
  CONSTRAINT "fk_match_actions_match" FOREIGN KEY ("match_id") REFERENCES "matches" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_match_actions_action_seq" CHECK (action_seq >= 1),
  CONSTRAINT "chk_match_actions_action_type" CHECK ((action_type)::text = ANY ((ARRAY['move'::character varying, 'wall'::character varying, 'resign'::character varying, 'timeout'::character varying, 'abort'::character varying])::text[])),
  CONSTRAINT "chk_match_actions_actor_seat" CHECK (actor_seat = ANY (ARRAY[0, 1]))
);
-- Create index "ux_match_actions_action_id" to table: "match_actions"
CREATE UNIQUE INDEX "ux_match_actions_action_id" ON "match_actions" ("match_id", "action_id");

-- ここから下は GORM のタグで表現できないぶんの手書き。

-- レーティングは負にならない。rating_after は進行中 NULL を許す形に張り替える。
ALTER TABLE "match_participants" DROP CONSTRAINT "chk_match_participants_rating";
ALTER TABLE "match_participants" ADD CONSTRAINT "chk_match_participants_rating"
  CHECK ("rating_before" >= 0 AND ("rating_after" IS NULL OR "rating_after" >= 0));

-- 決着済みの participants は勝敗とレーティングを必ず持つ。
-- outcome を NULL 許容にしたぶん、「片方だけ埋まっている」中途半端な行を弾く。
ALTER TABLE "match_participants" ADD CONSTRAINT "chk_match_participants_settled"
  CHECK (("outcome" IS NULL AND "rating_after" IS NULL) OR ("outcome" IS NOT NULL AND "rating_after" IS NOT NULL));
