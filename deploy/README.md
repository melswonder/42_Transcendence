# 無料での公開デプロイ（Render + Supabase）

料金ゼロ・クレジットカード不要で公開 URL を得る構成。
1 つのコンテナに Caddy + Next.js + Go を同居させ、**同一オリジン**で配信する
（Cookie 認証と WebSocket が構成変更なしでそのまま動く）。

```
ブラウザ ── https://<app>.onrender.com ──> Caddy(:PORT)
                                            ├─ /api/* → Go backend (:4000)
                                            └─ それ以外 → Next.js (:3000)
DB: Supabase の無料 Postgres
```

## 手順

### 1. Supabase（DB・無料）
1. https://supabase.com で無料アカウントを作り、プロジェクトを作成
2. Connect → **Session pooler** の接続文字列をコピーし、末尾に `?sslmode=require` を付ける
   （Direct connection は IPv6 のみのため Render からは繋がらない。Pooler を使う）
3. パスワードに記号がある場合は URL エンコードする（例: `?` → `%3F`）。さらに `&search_path=public` も付ける

### 2. Render（アプリ・無料）
1. https://render.com で無料アカウントを作成（GitHub 連携）
2. **New → Blueprint** → このリポジトリを選択（ルートの `render.yaml` が読まれる）
3. 環境変数を入力:
   | 変数 | 値 |
   |---|---|
   | `DATABASE_URL` | 手順 1 の URL（末尾は `?sslmode=require&search_path=public`） |
   | `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | ローカルと同じ値 |
   | `GOOGLE_REDIRECT_URL` | `https://<app>.onrender.com/api/auth/google/callback` |
   | `FRONTEND_URL` | `https://<app>.onrender.com` |
   | `NEXT_PUBLIC_API_URL` | `https://<app>.onrender.com/api` |

   `<app>` はサービス名（render.yaml では `quoridor`。取られていると suffix が付くので
   最初のデプロイ後に実際の URL を確認し、必要なら 3 変数を直して再デプロイする）

### 3. Google Cloud Console
承認済みリダイレクト URI に `GOOGLE_REDIRECT_URL` と同じ値を追加する。

## ローカルでの動作確認

```bash
docker build -f deploy/Dockerfile --build-arg NEXT_PUBLIC_API_URL=http://localhost:8080/api -t quoridor-deploy .
docker run --rm -p 8080:8080 -e PORT=8080 \
  -e DATABASE_URL="postgresql://postgres:postgres@host.docker.internal:5432/transcendence?sslmode=disable" \
  -e GOOGLE_CLIENT_ID=dummy -e GOOGLE_CLIENT_SECRET=dummy \
  -e GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback \
  -e FRONTEND_URL=http://localhost:8080 quoridor-deploy
# → http://localhost:8080
```

## 無料枠の注意点
- **15 分アクセスが無いとスリープ**し、次のアクセスは起動に 1 分ほどかかる（Render free）
- **アバター画像は再デプロイ/再起動で消える**（無料プランは永続ディスクなし）
- Supabase の無料プロジェクトは **1 週間ノーアクセスで一時停止**（ダッシュボードから再開）
- NEXT_PUBLIC_API_URL はビルド時に焼き込まれるため、URL を変えたら再デプロイが必要
