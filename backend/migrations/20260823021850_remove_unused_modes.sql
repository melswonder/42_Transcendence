-- Modify "matches" table
ALTER TABLE "matches" DROP CONSTRAINT "chk_matches_mode", ADD CONSTRAINT "chk_matches_mode" CHECK ((mode)::text = ANY ((ARRAY['ranked'::character varying, 'casual'::character varying])::text[]));
