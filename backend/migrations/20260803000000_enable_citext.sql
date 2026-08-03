-- citext は GORM のタグでは表現できないため手書きで補う（docs/database-design.md §13）。
-- users.email / oauth_accounts.provider_email が citext 型を使うので、
-- テーブル作成より先に実行される必要がある（ファイル名の時刻で順序を制御）。
CREATE EXTENSION IF NOT EXISTS citext;
