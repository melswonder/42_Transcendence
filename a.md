# モジュール解説書（ft_transcendence 評価用）

本プロジェクトが申告する **12 モジュール / 19 点**（Major 7 × 2pt + Minor 5 × 1pt）について、
「何を実装したか」「どのファイルにあるか」「誰が担当したか」「評価でどう実演するか」
「何を聞かれてどう答えるか」をモジュール単位でまとめたもの。

課題書の必要点数は 14 点。ボーナス上限が 5 点なので、**19 点は上限ちょうど**。

| # | カテゴリ | モジュール | 種別 | Pts | 担当 |
| --- | --- | --- | --- | --- | --- |
| [1](#1-web--フレームワークの使用frontend--backend) | IV.1 Web | フレームワークの使用（FE + BE） | Major | 2 | hirwatan |
| [2](#2-web--websocket-によるリアルタイム機能) | IV.1 Web | WebSocket によるリアルタイム機能 | Major | 2 | ttanaka |
| [3](#3-web--公開-apiキー--レート制限--ドキュメント--5エンドポイント以上) | IV.1 Web | 公開 API | Major | 2 | atashiro |
| [4](#4-gaming--web-ゲーム本体quoridor) | IV.6 Gaming | Web ゲーム本体（Quoridor） | Major | 2 | ttanaka |
| [5](#5-gaming--リモートプレイヤー) | IV.6 Gaming | リモートプレイヤー | Major | 2 | ttanaka |
| [6](#6-user-management--標準的なユーザー管理認証) | IV.3 User Mgmt | 標準的なユーザー管理・認証 | Major | 2 | sguruge |
| [7](#7-data--高度なアナリティクスダッシュボード) | IV.8 Data | 高度なアナリティクスダッシュボード | Major | 2 | atashiro |
| [8](#8-web--orm-の使用) | IV.1 Web | ORM の使用 | Minor | 1 | kanahash |
| [9](#9-user-management--ゲーム統計と対戦履歴) | IV.3 User Mgmt | ゲーム統計と対戦履歴 | Minor | 1 | sguruge |
| [10](#10-user-management--oauth-20-によるリモート認証) | IV.3 User Mgmt | OAuth 2.0 | Minor | 1 | hirwatan |
| [11](#11-gaming--観戦モード) | IV.6 Gaming | 観戦モード | Minor | 1 | sguruge |
| [12](#12-accessibility--多言語対応3-言語) | IV.2 A11y/i18n | 多言語対応（3 言語） | Minor | 1 | kanahash |

---

## 全体アーキテクチャ（どのモジュールの説明でも前提になる）

```
ブラウザ
  │  HTTPS / WSS
  ▼
Caddy (caddy/Caddyfile)        :443 / :8443   TLS 終端・ローカル CA
  ├─ /api/*  ──────────────►  Backend  (Go 1.26 + Gin)      :4000
  └─ その他  ──────────────►  Frontend (Next.js 16 / React 19) :3000
                                   │
                              PostgreSQL 16 :5432
```

バックエンドは Clean Architecture の 4 層。**内側は外側を import しない**のが唯一の規律。

| 層 | ディレクトリ | 役割 | 外部依存 |
| --- | --- | --- | --- |
| domain | `backend/domain/` | エンティティとルール（Quoridor、Elo、実績、バリデーション） | 標準ライブラリと uuid のみ。**HTTP も DB も知らない** |
| usecase | `backend/usecase/` | シナリオの進行 + リポジトリ **インターフェース定義** | domain のみ |
| infrastructure | `backend/infrastructure/` | GORM リポジトリ、OAuth クライアント、presence、レートリミッタ、ファイル保存 | usecase のインターフェースを実装 |
| handler | `backend/handler/` | HTTP / WebSocket / SSE の入出力と JSON 変換 | usecase を呼ぶだけ |

配線は composition root 1 箇所（`backend/cmd/serv/main.go:168-181`）に集約。
依存の向きが内向きなので、domain と usecase は DB もネットワークも無しで単体テストできる
（テスト関数 **51 本**、`cd backend && go test ./...` で全部通る）。

---

# Major モジュール

## 1. Web — フレームワークの使用（Frontend + Backend）

**種別**: Major（2pt） / **担当**: hirwatan

### 課題要件
フロントエンドフレームワークとバックエンドフレームワークを両方使う。

### 何を使ったか

| | 採用 | 選定理由 |
| --- | --- | --- |
| Frontend | **Next.js 16**（App Router / React 19） | ファイルベースルーティング。Server Component で認証とデータ取得をサーバー側に置けるので、セッション Cookie をクライアント JS に晒さずに済む |
| Backend | **Go 1.26 + Gin** | goroutine が WebSocket の同時接続に素直に合う。Gin はルート木（radix tree）とミドルウェアチェーンを提供 |

### 該当ファイル

**Backend（Gin）**
- `backend/cmd/serv/main.go` — composition root。ミドルウェア 3 本（`gin.Recovery()` / `accessLog()` / `corsMiddleware()`）を `NewRouter` へ渡す（`:178`）
- `backend/handler/handler.go:95-180` — `NewRouter`。全ルートを `router.Group` で階層化して登録
- `backend/handler/handler.go:84-92` — `wrapF` アダプタ

**Frontend（Next.js）**
- `frontend/app/` — 14 ルート（`/`, `login`, `signup`, `game`, `watch`, `friends`, `settings`, `users/[userId]`, `stats`, `matches`, `achievements`, `leaderboard`, `privacy`, `terms`）
- `frontend/app/layout.tsx` — ルートレイアウト（Mantine Provider + next-intl Provider）
- `frontend/components/` — 30 以上の共有コンポーネント
- `frontend/lib/auth.ts` — Server Component 側の `getCurrentUser()`

### 技術的な要点：`wrapF` アダプタ

Gin へ移行するとき、全ハンドラを `gin.HandlerFunc` に書き換える選択肢もあったが、
そうすると handler 層が Gin に依存してしまう。代わりに 9 行のアダプタを 1 枚挟んだ。

```go
// backend/handler/handler.go:86
func wrapF(h http.HandlerFunc, params ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        for _, p := range params {
            c.Request.SetPathValue(p, c.Param(p))   // Gin の :userId → r.PathValue("userId")
        }
        h(c.Writer, c.Request)
    }
}
```

結果、**各ハンドラは `net/http` の `http.HandlerFunc` のまま**で、Gin を一切 import していない。
Gin をやめても差し替えるのは `handler.go` の 1 ファイルだけ。

### Server / Client Component の分割

```
frontend/app/stats/page.tsx        ← Server Component（async、Cookie を直接読む）
  └─ getSummary/getTimeseries/getBreakdown を Promise.all で並列取得（:39-43）
      └─ <StatsFilters />          ← "use client"（URL クエリを操作）
      └─ <RatingChart />           ← "use client"（Recharts）
      └─ <ExportButtons />         ← "use client"（window.print）
```

### デモ手順
1. `make` → https://localhost
2. `docker compose logs backend` に `server listening on :4000`、1 リクエスト 1 行のアクセスログが出る
3. ブラウザの Network タブで、`/stats` の初回描画が **HTML として返っている**（Server Component）ことを見せる

### 想定質問
- **Q. Next.js のフルスタック機能を使えば 1 つで両方満たせるのでは？**
  A. あえて分けた。ゲームの状態はサーバーがメモリに持つ長命なセッションで、goroutine とチャネルで扱いたかった。Next.js の Route Handler はリクエスト単位で、WebSocket の常時接続には向かない。
- **Q. Gin を使って何が変わった？**
  A. ルート定義がグループ単位で読めるようになった（`/auth`, `/matches`, `/stats`, `/game`, `/users`, `/media`, `/friends`, `/apikeys`, `/v1`）。ミドルウェアの適用順も `main.go` の 1 行で見える。
- **Q. CORS で詰まった点は？**
  A. 資格情報（Cookie）付きのリクエストでは、ブラウザは `Access-Control-Allow-Headers: *` をワイルドカードとして展開せず、文字通り `*` というヘッダ名として扱う。preflight が要求してきたヘッダをそのまま返す実装にした（`main.go:45-50`）。

---

## 2. Web — WebSocket によるリアルタイム機能

**種別**: Major（2pt） / **担当**: ttanaka

### 課題要件
- クライアント間のリアルタイム更新
- 接続・切断を破綻なく扱う
- 効率的なメッセージ配信

### 設計の中心：サーバー権威型 + 全量スナップショット

クライアントは盤面を計算しない。**サーバーだけが `domain.Quoridor` を書き換え**、
結果を全量スナップショットとして配る。差分は送らない。

```go
// backend/usecase/game.go:129
// GameStateView は盤面のスナップショット。差分ではなく毎回この全量を送る。
// 遅延や取りこぼしがあっても、最新の 1 通で盤面が揃う。
```

これにより、イベントバッファが溢れて古いイベントを捨てても盤面は壊れない
（`gameEventBuffer = 32`、`usecase/game.go:27`）。

### 該当ファイル

| ファイル | 役割 |
| --- | --- |
| `backend/handler/game.go` | WebSocket の upgrade、JSON メッセージの解釈、keep-alive ping（25 秒間隔、`:21`） |
| `backend/usecase/game.go` | セッション管理の本体（991 行）。マッチング、盤面適用、タイマー、配信 |
| `backend/infrastructure/game.go` | `match_actions` への追記、進行中対局の検索 |
| `backend/infrastructure/eventhub.go` | 対局終了を SSE 購読者へ配る プロセス内 Pub/Sub |
| `frontend/components/use-game-socket.ts` | 接続・指数バックオフ再接続・メッセージ処理のフック |
| `frontend/components/game-screen.tsx` | 対局画面 |

### メッセージプロトコル

```
クライアント → サーバー  (handler/game.go:41)
  { type: "join_queue" }
  { type: "leave_queue" }
  { type: "watch",  matchId }
  { type: "unwatch" }
  { type: "action", actionId, expectedVersion, action: "move"|"wall"|"resign", row, col, orientation }

サーバー → クライアント  (handler/game.go:58)
  { type: "queued" }
  { type: "state", state: { ...盤面全量... } }
  { type: "opponent_disconnected", graceSeconds: 45 }
  { type: "opponent_reconnected" }
  { type: "error", code, message }
```

エラーは **コードで返し、文言はクライアント側で組み立てる**（i18n のため。モジュール 12 参照）。

### 並行性・整合性の 3 つの仕掛け

**(a) 楽観的バージョニング** — `expectedVersion` が現在の版数とズレていたら拒否
```go
// usecase/game.go:726
if in.Kind != domain.GameActionResign && in.ExpectedVersion != s.version {
    return false, domain.ErrStaleGameVersion   // 投了だけはいつでも受ける
}
```

**(b) 冪等キー** — クライアントが生成する `actionId` で再送を 1 回だけ適用
```go
// usecase/game.go:717
if _, ok := s.applied[in.ActionID]; ok {
    from.push(GameEvent{Type: GameEventState, State: s.viewLocked()})  // 現盤面を返すだけ
    return false, nil
}
```
DB 側も `UNIQUE(match_id, action_id)` で二重防御。競り合いで DB が先に弾いた場合は
`domain.ErrDuplicateGameAction` として同じ扱いに落とす（`:758-764`）。

**(c) 複製に適用してから差し替え** — 永続化に失敗したらメモリ上の盤面は動かさない
```go
// usecase/game.go:731
trial := s.game.Clone()          // 複製に適用してルール検証
... AppendAction(ctx, record)    // 永続化
s.game = trial                   // 成功してから差し替え
s.version++
```

### 切断・再接続の扱い

| イベント | 動作 | 実装 |
| --- | --- | --- |
| 対局者が切断 | 45 秒の猶予タイマーを起動し、相手へ `opponent_disconnected` を配信 | `usecase/game.go:666-694`（`detach`） |
| 猶予内に再接続 | `Connect` が進行中の対局を DB から探して自動復帰、最新盤面を送信 | `usecase/game.go:222`（`Connect`）→ `findOrRestore`（`:244`） |
| 猶予切れ | 切断側の不戦敗として決着 | `onGraceExpired`（`:808`） |
| 手番の持ち時間切れ（60 秒） | 手番側の時間切れ負け | `onTurnTimeout`（`:800`） |
| サーバー再起動後の復帰 | `match_actions` を頭から再生して盤面を復元 | `buildSession`（`:272`）→ `replayAction`（`:295`） |

`TurnTimeLimit` / `ReconnectGrace` はテストで短縮できるよう `var`（`usecase/game.go:18-23`）。

### 効率的な配信

- 配信先は `s.subs` マップ（対局者 + 観戦者）。`broadcastExcept`（`:953`）で 1 回のループで配る
- 1 接続あたりバッファ 32。溢れたら**最も古いイベントを捨てる**（ブロックさせない）
- 決着済みセッションには配信しない（`closedForBroadcast`、`:697`）
- keep-alive ping は 25 秒間隔。中継機器のアイドル切断（多くは 30〜60 秒）より短くしてある

### テスト
`backend/usecase/game_test.go` にテスト関数 **15 本**。マッチング、冪等性、版数ズレ、
猶予切れ、時間切れ、観戦者の出入りをすべてカバー。

### デモ手順
1. 別ブラウザ（または片方シークレット）で 2 アカウントログイン → 両方 `/game` → クイックマッチ
2. 片方で駒を動かす → もう片方に即反映
3. **片方のタブを閉じる** → 残った側に「相手が切断（45 秒）」が出る
4. 45 秒以内に開き直す → 盤面が復元されて対局続行
5. もう一度閉じて 45 秒待つ → 残った側の勝ちで決着
6. `docker compose restart backend` してから再接続 → `match_actions` の再生で盤面が戻る

### 想定質問
- **Q. 差分ではなく全量を送るのは無駄では？**
  A. 1 スナップショットは壁が最大 20 枚 + 駒 2 個で数百バイト。Quoridor は 1 手あたり数秒に 1 回しか動かないので帯域は問題にならない。代わりに「取りこぼしたら壊れる」状態を作らずに済み、再接続時の復帰処理も同じコードで済む。
- **Q. `expectedVersion` と `actionId` は役割が被らない？**
  A. 別物。`expectedVersion` は「古い盤面を見て指した手」を弾く（相手が先に指していた場合）。`actionId` は「同じ手が 2 回届いた」を弾く（再送・二重クリック）。前者はユーザーの認識ズレ、後者は通信の重複。
- **Q. 複数サーバーにスケールしたら？**
  A. 現状 `gameSession` はプロセス内メモリなので 1 インスタンス前提。EventHub と PresenceHub も同じ。スケールするなら Redis Pub/Sub に置き換える必要がある（コード中にコメントで明記済み: `infrastructure/eventhub.go:21`）。

---

## 3. Web — 公開 API（キー / レート制限 / ドキュメント / 5エンドポイント以上）

**種別**: Major（2pt） / **担当**: atashiro

### 課題要件
- APIキーで保護
- レート制限
- ドキュメント
- GET / POST / PUT / DELETE を含む 5 エンドポイント以上

### エンドポイント一覧（`/v1` 配下、8 本）

| メソッド | パス | 必要スコープ | 内容 |
| --- | --- | --- | --- |
| GET | `/v1/profile` | read | 自分のプロフィール |
| **PUT** | `/v1/profile` | write | 表示名 / handle / 言語の更新 |
| GET | `/v1/matches` | read | 対戦履歴（from/to/mode/outcome/limit/offset で絞り込み） |
| GET | `/v1/stats` | read | 統計サマリ |
| GET | `/v1/leaderboard` | read | ランキング |
| GET | `/v1/friends` | read | フレンド一覧 |
| **POST** | `/v1/friends/requests` | write | フレンド申請 |
| **DELETE** | `/v1/friends/{userId}` | write | フレンド解除 |

要件の 4 メソッドすべてを含み、8 本で 5 本以上を満たす。
登録は `backend/handler/public.go:95-107`（`PublicHandler.Register`）。

### 該当ファイル

| ファイル | 役割 |
| --- | --- |
| `backend/handler/public.go` | `/v1` の HTTP 入口。`withKey` ミドルウェアと 8 ハンドラ |
| `backend/domain/apikey.go` | キーの生成・ハッシュ・スコープ判定・バリデーション |
| `backend/usecase/apikey.go` | 発行 / 一覧 / 失効 / 認証 / 認可 |
| `backend/infrastructure/apikey.go` | `api_keys` の永続化（`:17-103`）+ `FixedWindowLimiter`（`:105-147`） |
| `backend/handler/apikey.go` | 画面用のキー管理 API（`/apikeys`）。Cookie 認証 |
| `backend/apispec/` | OpenAPI アノテーションの置き場（11 ファイル） |
| `backend/docs/swagger/` | `swag init` が生成した spec |

### APIキーの設計

```go
// domain/apikey.go:63
func NewAPIKeySecret() string {
    return apiKeyPrefix + strings.ToLower(rand.Text()+rand.Text())   // "tsc_" + crypto/rand 約208bit
}

// domain/apikey.go:71
func HashAPIKey(raw string) string {
    sum := sha256.Sum256([]byte(raw))      // DB には SHA-256 だけ保存
    return hex.EncodeToString(sum[:])
}
```

| 性質 | 実装 |
| --- | --- |
| 生キーは一度だけ表示 | 作成レスポンスにのみ含まれる。DB には `key_hash CHAR(64) UNIQUE` だけ |
| 見分け用の接頭辞 | `tsc_` + 先頭 6 文字を `key_prefix` に保存（一覧表示用、`APIKeyPrefixOf`） |
| スコープ | `read` / `write`。エンドポイントごとに要求（`HasScope`） |
| 有効期限 | `expires_at`（nil なら無期限） |
| 失効 | `revoked_at`。以後そのキーは 401 |
| 最終使用 | `last_used_at` を認証のたびに更新（失敗は握りつぶす、`TouchKey`） |

**bcrypt を使わない理由**: キーは 208bit の暗号学的乱数で、パスワードと違い辞書攻撃・
総当たりが現実的でない。一方で認証は毎リクエスト走るので、意図的に遅いハッシュは害になる。
同じ判断をセッショントークンにも適用している（`domain/auth.go:56-58`）。

### 関門の順序（`withKey`、`handler/public.go:47`）

```
Authorization: Bearer <key> を取り出す
  └─ 無い          → 401 missing_api_key
Authenticate（ハッシュ照合）
  ├─ 失効          → 401 api_key_revoked
  ├─ 期限切れ      → 401 api_key_expired
  └─ 不一致        → 401 invalid_api_key
Authorize（スコープ + 流量）
  ├─ スコープ不足  → 403 insufficient_scope
  └─ 流量超過      → 429 rate_limited  + Retry-After
ハンドラ本体
```

**失敗の理由ごとにコードを分けている**のが要点。「なぜ弾かれたか」がクライアントに伝わる。

### レート制限

固定窓 1 分、キー単位（`FixedWindowLimiter`、`infrastructure/apikey.go:105`）。
既定 60 req/min、環境変数 `PUBLIC_API_RATE_LIMIT` で変更可
（`cmd/serv/main.go:130` — **デモで下げられるようにしてある**）。

**成功時にもヘッダを返す**のがポイント。curl だけで「あと何回で 429 か」を実演できる。

```
X-RateLimit-Limit:     60
X-RateLimit-Remaining: 57
X-RateLimit-Reset:     1757671234        (Unix秒)
Retry-After:           38                (429 のときだけ)
```

### ドキュメント

Swagger UI を `GET /swagger/index.html` で配信（`handler/handler.go:106`）。
`backend/apispec/` のアノテーションから `swag init` で生成し、`docs/swagger/` にコミット。

### デモ手順

```bash
# 1. 画面（/settings）でキーを作成し、一度だけ表示される tsc_... を控える
KEY=tsc_xxxxxxxx

# 2. GET（read スコープ）
curl -k -i https://localhost/api/v1/profile -H "Authorization: Bearer $KEY"
#   → 200 + X-RateLimit-Remaining: 59

# 3. PUT（write スコープ）
curl -k -X PUT https://localhost/api/v1/profile \
  -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"display_name":"demo"}'

# 4. read だけのキーで write を叩く → 403 insufficient_scope
curl -k -i -X PUT https://localhost/api/v1/profile -H "Authorization: Bearer $READ_ONLY_KEY" -d '{}'

# 5. レート制限（先に PUBLIC_API_RATE_LIMIT=5 で起動しておくと速い）
for i in $(seq 1 8); do
  curl -sk -o /dev/null -w '%{http_code} ' https://localhost/api/v1/stats -H "Authorization: Bearer $KEY"
done
#   → 200 200 200 200 200 429 429 429

# 6. キーを失効させてから叩く → 401 api_key_revoked
# 7. でたらめなキー → 401 invalid_api_key
```

Swagger UI: https://localhost/api/swagger/index.html

### 既知の差分（正直に申告すべき点）

OpenAPI spec（`docs/swagger/swagger.json`、39 パス）には、**現在ルーターに登録されていない
エンドポイントが含まれている**:

- `/blocks`（GET/POST/DELETE）、`/auth/refresh`、`/auth/sessions`、`/auth/oauth/{provider}`、
  `/auth/oauth/accounts`、`DELETE /users/me`
- 逆に `/game/ws`、`/game/live`、`/stats/opponents` は spec に無い

`/v1` の 8 本は spec とルーターが完全に一致しているので、**本モジュールの要件（公開APIの
ドキュメント）自体は満たしている**が、Swagger UI の "Try it out" で `/blocks` を叩かれると
404 になる。評価前に `apispec/blocks.go` などを削除するか、実装を追加して揃えておくのが望ましい。

### 想定質問
- **Q. なぜキー管理 API（`/apikeys`）を `/v1` に入れない？**
  A. キーでキーを発行できると失効の意味が薄れる。キーのライフサイクルは Cookie セッション（= 本人がブラウザでログイン）でのみ操作できるようにした（`handler/public.go:93-94` にコメント）。
- **Q. 固定窓だとバーストが通る（窓の境界で 2 倍）が？**
  A. 承知のうえ。スライディングウィンドウやトークンバケットに比べて状態が 1 バケット（開始時刻 + カウント）で済み、`X-RateLimit-Reset` をそのまま返せる。攻撃対策ではなく公平な流量配分が目的なので、この粒度で足りると判断した。
- **Q. レート制限がメモリなのでサーバーを再起動すると消える**
  A. 消える。複数インスタンス化するときは Redis に移す前提でコメントを残してある（`infrastructure/apikey.go:106`）。
- **Q. `/v1/profile` の PUT でアバターを更新できないのはなぜ？**
  A. 画像アップロードが multipart + Cookie セッション前提の実装で、キー認証の経路に載せていない（`handler/public.go:125` にコメント）。

---

## 4. Gaming — Web ゲーム本体（Quoridor）

**種別**: Major（2pt） / **担当**: ttanaka

### 課題要件
- ユーザー同士が対戦できる Web ゲーム
- ライブ対戦
- 明確なルールと勝敗条件

### なぜ Quoridor か

Pong が例示されているが、あえて **Quoridor**（コリドール）にした。
ターン制なので毎フレームの同期が要らず、代わりに**ルール検証がサーバーでしかできない**
（壁が相手の道を完全に塞いでいないかを BFS で調べる必要がある）。
サーバー権威型の設計が本質的に必要になるゲームを選んだ。

### ルール

- 9×9 マス、壁のアンカーは 8×8 の交点
- 各プレイヤー 壁 10 枚
- seat 0 は `(0,4)` から始まり **Row 8** へ、seat 1 は `(8,4)` から **Row 0** へ
- 1 手につき「駒を動かす」か「壁を置く」のどちらか
- 相手の駒には乗れない → 正面ジャンプ、塞がれていたらサイドステップ
- **どちらのゴール経路も完全に塞ぐ壁は置けない**（公式ルール）
- 勝敗: ゴール到達 / 投了 / 時間切れ（60 秒）/ 切断放棄（45 秒）

### 該当ファイル

| ファイル | 行数 | 役割 |
| --- | --- | --- |
| `backend/domain/quoridor.go` | 321 | **ルールエンジン本体**。HTTP も DB も import しない純粋な Go |
| `backend/domain/quoridor_test.go` | 424 | テーブル駆動テスト 9 本 |
| `backend/usecase/game.go` | 991 | セッション管理（モジュール 2 参照） |
| `frontend/components/quoridor-board.tsx` | | 盤面描画・クリック判定 |
| `frontend/components/game-screen.tsx` | | 対局画面全体 |
| `frontend/components/player-panel.tsx` | | プレイヤー情報と残り時間 |

### ルールエンジンの中身

**駒の移動（`LegalPawnMoves`、`quoridor.go:156`）**
```go
for _, d := range directions() {            // 上下左右
    next := 隣のマス
    if 盤外 || 壁で塞がれている { continue }
    if next != 相手の駒 { 合法手に追加; continue }
    // 相手の駒の上には乗れない
    jump := さらに 1 マス先
    if jump が盤内 && 塞がれていない { 正面ジャンプを追加; continue }
    // ジャンプできないときだけ、相手駒の左右へサイドステップ
    for _, p := range perpendicular(d) { ... }
}
```

**壁の重なり判定（`wallsConflict`、`quoridor.go:226`）** — 3 パターンを区別
1. 完全に同じ位置
2. **同じ向きで 1 マスずれた壁**（壁は 2 マス分の長さなので半分重なる）
3. **向きが違い、同じアンカーを共有**（十字に交差する）

**ゴール経路の保証（`hasPathToGoal`、`quoridor.go:275`）** — 幅優先探索
```go
trial := append(slices.Clone(q.Walls), w)     // 置いてみて
for seat := range quoridorSeats {
    if !hasPathToGoal(q.Pawns[seat], GoalRow(seat), trial) {
        return ErrWallSealsGoal                // どちらか一方でも塞がるなら却下
    }
}
```
9×9 = 81 マスの BFS なので、`[9][9]bool` の訪問済み配列で十分。ヒープ確保も無し。

### エラーの設計

ドメインエラーを 10 種類定義（`quoridor.go:27-36`）:
`ErrNotYourTurn` / `ErrGameFinished` / `ErrCellOutOfBoard` / `ErrIllegalPawnMove` /
`ErrWallOutOfBoard` / `ErrInvalidWall` / `ErrNoWallsLeft` / `ErrWallOverlaps` /
`ErrWallSealsGoal` / `ErrInvalidSeat`

usecase が `errors.Is` で分岐し、handler が HTTP ステータス / WS エラーコードに翻訳する。
**domain 層はエラーの「意味」だけを定義し、HTTP を知らない。**

### 不変条件：局面を壊さない

すべての操作は **「失敗したら局面を一切変えない」** ことを保証している。

```go
// quoridor.go:126  PlaceWall
if q.WallsLeft[seat] <= 0 { return ErrNoWallsLeft }        // 検証を全部先に
if err := q.validateWall(w); err != nil { return err }
q.Walls = append(q.Walls, w)                                // 通ってから初めて変更
q.WallsLeft[seat]--
```

さらに usecase 側が `Clone()` してから適用するので、二重に守られている。

### デモ手順
1. 2 アカウントでクイックマッチ
2. 駒を動かす／壁を置く。**合法手がハイライトされる**（`legalMoves` をサーバーが計算して送っている）
3. **相手の駒の正面に自駒を移動 → ジャンプできることを見せる**
4. その先を壁で塞いでから接近 → **サイドステップになる**
5. 相手のゴール手前を壁で完全に塞ごうとする → **「相手の経路を塞ぐ壁は置けません」で拒否される**
6. 壁を 10 枚使い切る → 11 枚目が拒否される
7. 投了ボタン → 相手の勝ちで決着、結果モーダル
8. `cd backend && go test ./domain/ -run Quoridor -v` でルールのテストを見せる

### 想定質問
- **Q. クライアントでルール検証していないのか？**
  A. していない。クライアントはサーバーが送ってきた `legalMoves` を描画するだけ。**改造クライアントで不正な手を送っても `domain` 側で弾かれる**。これが「サーバー権威型」の意味。
- **Q. BFS は毎回全探索で重くないか？**
  A. 壁を置くときだけ、2 回（両プレイヤー分）走る。81 マスなので最悪 81 ノード × 4 方向。1 手あたり数マイクロ秒。壁を置く操作自体が 1 局に最大 20 回しかない。
- **Q. 引き分けはあるか？**
  A. Quoridor のルール上は必ず決着する（壁で完全に塞げないので、必ずゴールへの経路が残る）。ただし DB の `match_participants.outcome` は `draw` を許容し、Elo も `score=0.5` を扱える（`domain/rating.go:46`）。中断（`abort`）のときは勝敗を付けない。

---

## 5. Gaming — リモートプレイヤー

**種別**: Major（2pt） / **担当**: ttanaka

### 課題要件
- 別のコンピュータにいる 2 人が同じゲームをリアルタイムで対戦
- ネットワーク遅延と切断を破綻なく扱う
- 再接続ロジック

> **モジュール 2（リアルタイム機能）との違い**
> モジュール 2 は「WebSocket という技術基盤」、本モジュールは「別 PC の 2 人が対戦として成立する」こと。
> 同じコードベースを使うが、要求される性質が違う（前者は配信の効率と接続管理、後者は遅延下での対戦の公平性）。

### 別 PC で成立するための要素

| 要素 | 実装 |
| --- | --- |
| マッチメイキング | クイックマッチ待機列。相手が来たら即座に対戦を作る（`usecase/game.go:355` `JoinQueue` → `:402` `startMatch`） |
| 座席の割り当て | 先に並んでいた方が seat 0（先手）。`startMatch(first, second)` |
| 盤面の向き | **プレイヤーごとに自分が手前になるよう反転して描画**（`frontend/components/quoridor-board.tsx`） |
| 遅延への耐性 | 全量スナップショット。遅れて届いても最新の 1 通で揃う |
| 遅い操作の拒否 | `expectedVersion` による楽観ロック |
| 再送の重複防止 | `actionId` による冪等性 |
| 切断猶予 | 45 秒（`ReconnectGrace`） |
| 自動再接続 | クライアント側で指数バックオフ（上限 10 秒、`use-game-socket.ts:59`） |
| 自動復帰 | 再接続時にサーバーが進行中の対局を探して自動で座らせる |
| 手番の持ち時間 | 60 秒（`TurnTimeLimit`）。切れたら時間切れ負け |

### 再接続のシーケンス

```
プレイヤーB の回線が切れる
  │
  ├─ サーバー: detach() → conns[seatB] == 0 → 45秒タイマー起動
  │            → プレイヤーA へ {opponent_disconnected, graceSeconds:45}
  │
  ├─ クライアントB: onclose → 指数バックオフで再接続（1s, 2s, 4s, ... 最大10s）
  │
  ├─【猶予内に繋がった場合】
  │    サーバー: Connect() → findOrRestore() が FindActiveMatch で進行中対局を発見
  │              → 既存 gameSession に attach → 猶予タイマーを止める
  │              → B へ最新 state、A へ {opponent_reconnected}
  │    → 対局続行
  │
  └─【45 秒を超えた場合】
       サーバー: onGraceExpired(seatB) → forceFinish
                 → B の負けで決着（resultType: timeout）
                 → A へ最終 state（ratingAfter 付き）
```

サーバー自体が落ちた場合も、`match_actions`（append-only の手のログ）を
`replayAction` で頭から再生して盤面を復元する（`usecase/game.go:272-317`）。

### 該当ファイル
- `backend/usecase/game.go:355-423` — 待機列と対戦開始
- `backend/usecase/game.go:222-294` — 接続と復帰（`Connect` / `findOrRestore` / `buildSession`）
- `backend/usecase/game.go:666-694` — 切断（`detach`）と猶予タイマー
- `backend/usecase/game.go:800-873` — タイムアウト系（`onTurnTimeout` / `onGraceExpired` / `forceFinish`）
- `frontend/components/use-game-socket.ts` — 再接続フック
- `backend/usecase/game_test.go` — テスト 15 本

### デモ手順
**同じ LAN の別マシン 2 台で行うのが理想**（無理なら別ブラウザ + シークレットウィンドウ）
1. 両方でログイン → `/game` → クイックマッチ → 対戦成立
2. 盤面が**互いに自分が手前**になっていることを見せる
3. 交互に指す。両画面が即座に揃う
4. 片方の **Wi-Fi を切る**（または開発者ツールの Network を Offline に）
   → 相手画面に「相手が切断しました（45 秒）」+ カウントダウン
5. 20 秒後に回線を戻す → **自動で再接続し、盤面が復元されて続行**
6. もう一度切って 45 秒放置 → 残った側の勝ちで決着、レーティングが動く
7. 手番のまま 60 秒放置 → 時間切れ負け

### 想定質問
- **Q. 遅延補償（クライアント予測・ロールバック）は？**
  A. 入れていない。Quoridor はターン制で、1 手あたり数秒〜数十秒の思考時間がある。100ms の遅延はゲーム体験に影響しない。代わりに予測を入れないことで「クライアントとサーバーで盤面が食い違う」クラスのバグが構造的に発生しない。
- **Q. 45 秒という数字の根拠は？**
  A. モバイル回線の切り替え（Wi-Fi ↔ 4G）やスリープ復帰が数秒〜十数秒。それを吸収しつつ、待たされる側が我慢できる上限として設定。`var` なのでテストでは短縮している。
- **Q. 両方同時に切断したら？**
  A. 両者に猶予タイマーが立つ。先に切れた方が先に負け判定される。`forceFinish` は `s.game.Finished()` を見るので二重決着しない（`usecase/game.go:822`）。

---

## 6. User Management — 標準的なユーザー管理・認証

**種別**: Major（2pt） / **担当**: sguruge（認証部分は hirwatan）

### 課題要件
- プロフィール情報を更新できる
- アバターをアップロードできる（未設定時はデフォルト）
- 他ユーザーをフレンドに追加し、オンライン状態が見られる
- 情報を表示するプロフィールページがある

### 4 要件の対応表

| 要件 | 実装 | ファイル |
| --- | --- | --- |
| プロフィール更新 | 表示名（50字）/ handle（30字、ユニーク）/ 言語 | `handler/user.go` `PATCH /users/me`、`usecase/user.go` |
| アバター | png/jpeg/webp、5MB まで。未設定はデフォルト画像 | `usecase/media.go`、`handler/media.go` |
| フレンド + オンライン状態 | 申請 / 承認 / 拒否 / 解除、2 分ウィンドウのオンライン判定 | `usecase/friend.go`、`infrastructure/presence.go` |
| プロフィールページ | 自分と他人の両方 | `frontend/app/settings/page.tsx`、`frontend/app/users/[userId]/page.tsx` |

### 認証の仕組み

```
メール+パスワード                    Google OAuth 2.0（モジュール 10）
  │ bcrypt（salt 込み）                │
  └──────────┬───────────────────────┘
             ▼
    セッショントークン発行
      生トークン → HttpOnly Cookie（クライアントへ）
      SHA-256   → sessions.token_hash（DB へ）
      有効期限 7 日
```

**DB には生トークンを一切保存しない**（`domain/auth.go:33-62`）。
DB が漏れても、その中身だけでは誰にもなりすませない。

Cookie 属性: `HttpOnly` / `Secure`（HTTPS 時）/ `SameSite=Lax`（`handler/auth.go:244-247`）。
`Strict` にしない理由は Google からのリダイレクトで Cookie が送られなくなるため（コメントに明記）。

**アカウントの存在を悟らせない**（`usecase/auth.go:166-180`）:
ユーザーが存在しない / OAuth 専用でパスワードが無い / パスワード不一致 —
この 3 つをすべて同じエラーに丸め、**ダミー照合で応答時間も揃える**。

### アバターの検証（`usecase/media.go:55-106`）

申告された `Content-Type` や拡張子は信用しない。3 段階で検証:

```
1. サイズ上限        handler の MaxBytesReader（第一関門）+ usecase 側でも再確認（5MB）
2. 中身から MIME 判定  先頭バイトを sniff → image/png, image/jpeg, image/webp のみ許可
3. デコード確認       実際に画像としてデコードできるか + 寸法を取得
```

保存名は推測できないランダム値（`newStorageKey`、`:137`）。
`storage_key` は **API から一切外に出さない** → URL 総当たりで他人の画像を列挙できない。

削除は論理削除（`status='deleted'`）。使用中のアバターなら同一トランザクションで
`users.avatar_asset_id` を外し、デフォルトに戻る。ファイル本体は消さない
（配信は `status` を見て止まるので急がなくてよい）。

### フレンドシステム

**関係の正規化**: `friendships` は `(user_low_id, user_high_id)` の複合主キーで
`CHECK(user_low_id < user_high_id)`。**同じ 2 人の関係が 2 行できない**ことを DB が保証。
誰が申請したかは `requested_by_user_id` で別に持つ。

**相互申請の自動承認**（`usecase/friend.go:82-134`）:
A が B に申請 → B が A に申請 → その場で `accepted` になる。
「自分が申請したのに相手からも申請が来ていて承認待ち」という不自然な状態を作らない。

**オンライン状態**（`infrastructure/presence.go`）:
```go
const PresenceOnlineWindow = 2 * time.Minute   // usecase/friend.go:15
```
`PresenceHub` がメモリで「最後に見かけた時刻」を持つ。
認証付きリクエストが通るたびに `Touch`（`handler/handler.go:60-67` で認証関数をラップ）。
対局中は WebSocket の keep-alive（25 秒間隔）が活動になる。

**DB に書かない理由**: 毎リクエスト `UPDATE users SET last_seen_at = now()` が走ると
書き込み負荷が対戦記録と競合する。presence は数分で埋め直される揮発情報なので、
再起動で消えても実害がない（`presence.go:12-16` にコメント）。

### 該当ファイル

| ファイル | 役割 |
| --- | --- |
| `backend/domain/user.go` | ユーザーのバリデーション（表示名・handle・ロケール） |
| `backend/domain/auth.go` | セッション / トークン生成・ハッシュ / OAuth state・nonce |
| `backend/domain/media.go` | アバターの制限値とエラー定義 |
| `backend/domain/friend.go` | フレンド関係の値オブジェクト |
| `backend/usecase/auth.go` | 登録 / ログイン / セッション / ログアウト |
| `backend/usecase/user.go` | プロフィールの取得・更新 |
| `backend/usecase/media.go` | アバターの検証・保存・配信・削除 |
| `backend/usecase/friend.go` | 申請 / 承認 / 解除 / オンライン判定 |
| `backend/infrastructure/presence.go` | メモリ上の presence |
| `backend/infrastructure/auth.go` | セッションと OAuth の永続化、Google クライアント |
| `frontend/app/settings/page.tsx` | 自分のプロフィール編集 |
| `frontend/app/users/[userId]/page.tsx` | 他人のプロフィール |
| `frontend/app/friends/page.tsx` | フレンド画面 |
| `frontend/components/user-avatar.tsx` | アバター表示（デフォルトのフォールバック込み） |

### デモ手順
1. `/signup` でメール+パスワード登録 → 自動ログイン
2. `/settings` で表示名と handle を変更
3. 既に使われている handle を入れる → **409 handle_taken**
4. アバターに png をアップロード → 反映。削除するとデフォルトに戻る
5. **拡張子を `.png` に変えただけのテキストファイル**をアップロード → **415 で拒否**（中身を見ている証拠）
6. 6MB の画像 → **413 で拒否**
7. 2 アカウントでフレンド申請 → 承認 → 一覧にオンライン表示
8. 片方をログアウトして 2 分待つ → オフライン表示に変わる
9. **A→B、B→A の順に申請** → 承認操作なしで即 accepted になる
10. `cd backend && go test ./usecase/ -run 'Friend|Media|Password' -v`

### 想定質問
- **Q. JWT ではなく不透明なセッショントークンにした理由は？**
  A. 即座に失効させられるから。JWT は署名が有効な限りサーバー側で止められず、ブラックリストを持つなら結局 DB を引くことになる。それなら最初からセッションテーブルを引けばいい。
- **Q. presence が再起動で消えるのは問題では？**
  A. 消えるが、次のリクエストで即座に `Touch` されるので最大 1 リクエスト分しかズレない。オンライン表示は本質的に近似値なので許容した。永続化のコストのほうが大きい。
- **Q. `user_low_id < user_high_id` の正規化は何のため？**
  A. アプリのバグで (A,B) と (B,A) の 2 行ができるのを **DB レベルで不可能にする**ため。フレンド関係は無向なので、方向を持たせる理由がない。

---

## 7. Data — 高度なアナリティクスダッシュボード

**種別**: Major（2pt） / **担当**: atashiro（統計集計は sguruge）

### 課題要件
- インタラクティブなチャート（折れ線・棒・円など）
- リアルタイムのデータ更新
- エクスポート機能（PDF、CSV など）
- カスタマイズ可能な日付範囲とフィルタ

### 4 要件の対応表

| 要件 | 実装 |
| --- | --- |
| インタラクティブなチャート | 折れ線（レーティング推移）/ 棒（対戦数）/ 円（勝敗内訳）— すべてホバーでツールチップ |
| リアルタイム更新 | SSE `GET /stats/stream` で `match_recorded` を受けたら `router.refresh()` |
| エクスポート | CSV（サーバー生成）/ PDF（印刷用 CSS + `window.print()`） |
| フィルタ | 期間（from/to）/ モード / 勝敗 / 対戦相手 の 4 軸 + 時系列の刻み（日/週/月） |

### 該当ファイル

**Frontend**
| ファイル | 役割 |
| --- | --- |
| `frontend/app/stats/page.tsx` | ダッシュボード（Server Component）。3 API を `Promise.all` で並列取得（`:39-43`） |
| `frontend/components/stats-filters.tsx` | フィルタ UI。**状態を URL のクエリに持つ** |
| `frontend/components/rating-chart.tsx` | 折れ線（レーティング推移） |
| `frontend/components/activity-chart.tsx` | 棒（対戦数の推移） |
| `frontend/components/outcome-chart.tsx` | 円（勝敗内訳） |
| `frontend/components/stat-tile.tsx` | 数値タイル（勝率・連勝・順位・レベル） |
| `frontend/components/export-buttons.tsx` | CSV / PDF ボタン |
| `frontend/components/use-stats-stream.ts` | SSE 購読フック |
| `frontend/lib/stats.ts` | 統計 API のクライアント |

**Backend**
| ファイル | 役割 |
| --- | --- |
| `backend/handler/stats.go` | 統計 API 5 本 |
| `backend/handler/match.go:162-236` | CSV エクスポート |
| `backend/handler/query.go` | フィルタのパースと検証 |
| `backend/usecase/stats.go` | 集計の組み立て |
| `backend/infrastructure/stats.go` | 集計 SQL（375 行） |
| `backend/domain/stats.go` | 勝率・連勝の計算、刻みの正規化 |
| `backend/infrastructure/eventhub.go` | 対局終了イベントの配信 |
| `backend/handler/achievement.go:Stream` | SSE エンドポイント |

### API

| エンドポイント | 内容 |
| --- | --- |
| `GET /stats/me` | サマリ（勝敗・勝率・連勝・レーティング・順位・レベル・XP） |
| `GET /stats/me/timeseries` | 時系列（`interval=day\|week\|month`） |
| `GET /stats/me/breakdown` | 内訳（円グラフ用） |
| `GET /stats/opponents` | **対戦したことのある相手だけ**を返す（フィルタの選択肢用） |
| `GET /stats/stream` | SSE。対戦が記録されたら `match_recorded` |
| `GET /leaderboard` | ランキング |
| `GET /matches/export.csv` | CSV エクスポート |

### 設計判断 1：フィルタの状態を URL クエリに持つ

```ts
// frontend/components/stats-filters.tsx:19-22
/** 絞り込みは URL のクエリに持つ。
 * Server Component が searchParams から読んでそのまま backend に渡せるので
 * 状態管理を増やさずに済み、共有やリロードにも耐える。
 */
```

副次効果: **フィルタした状態の URL をそのまま共有・ブックマークできる**。
`page.tsx` は `searchParams` を受け取ってバックエンドへ素通しするだけなので、
フロントに状態管理ライブラリが要らない。

### 設計判断 2：SSE のイベントに値を載せない

```ts
// frontend/components/use-stats-stream.ts:8-12
/** イベントには統計そのものが載っていない（載せると「イベントで運ばれた値」と
 * 「API で取り直した値」の 2 経路ができて食い違う）。合図として使い、
 * router.refresh() で Server Component 側にもう一度引かせる。
 */
```

イベントは「何かが起きた」という合図だけ。データの取得経路は常に 1 本に保つ。
これで「リアルタイム表示とリロード後の表示が違う」というバグが原理的に起きない。

**WebSocket ではなく SSE にした理由**: サーバー → クライアントの単方向でよく、
ブラウザの `EventSource` が再接続を自動でやってくれる（`use-stats-stream.ts:26-28`）。

### 設計判断 3：CSV はサーバーで生成、PDF はブラウザの印刷

| | 実装 | 理由 |
| --- | --- | --- |
| CSV | `GET /matches/export.csv`（`handler/match.go:167`） | 画面に出ている 1 ページ分ではなく、**条件に合う全件**を書き出したい。ページングだけ無視して同じフィルタを適用。無制限だとメモリを使い切れるので `csvExportLimit` で蓋をしてある |
| PDF | `window.print()` + 印刷用 CSS | チャートの見た目をそのまま出したい。サーバーで PDF を作ると Chromium を同梱するか、チャートを再実装することになる。印刷用 CSS でサイドバーと操作ボタンを隠している（`globals.css` の `print:hidden`） |

**エクスポートは常に現在のフィルタと一致する**（`export-buttons.tsx:20-24` が
`from/to/mode/outcome/opponent` を URL からそのまま引き継ぐ）。

### 連勝の計算を SQL でやらない理由

```go
// domain/stats.go:94-96
// SQL の窓関数でも書けるが、連続の判定はクエリが読みにくくなるうえ
// テストしづらいので、勝敗の列だけ取ってきてここで数える。
func Streaks(outcomes []string) (current, best int) { ... }
```
`domain/stats_test.go` に単体テストがある。

### 勝率に引き分けを含める

```go
// domain/stats.go:25-26
// WinRate は勝率 (0.0-1.0)。引き分けも母数に含める。
// 「勝ちでも負けでもない試合」を無かったことにすると、
// 引き分けの多い人の勝率が実態より高く出るため。
```

### デモ手順
1. 数局プレイして履歴を作る
2. `/stats` を開く → 折れ線・棒・円 + 数値タイル
3. 各チャートにホバー → ツールチップ
4. 期間フィルタを変える → **URL が変わり、全チャートが同時に更新される**
5. 対戦相手フィルタ → 対戦したことのある人しか選択肢に出ない
6. 刻みを日→週→月に変える
7. **URL をコピーして別タブで開く** → 同じフィルタ状態が復元される
8. **`/stats` を開いたまま、別ブラウザで対戦を 1 局終わらせる**
   → **リロードせずにチャートが更新される**（SSE）
9. CSV ボタン → ダウンロードした CSV の行数がフィルタ結果と一致することを見せる
10. PDF ボタン → 印刷プレビューでサイドバーとボタンが消えている

### 想定質問
- **Q. なぜ PDF をサーバーで生成しない？**
  A. チャートは Recharts が SVG で描いているので、サーバーで同じ絵を出すにはヘッドレスブラウザを動かすか、チャートを 2 度実装することになる。印刷用 CSS なら数行で、しかも見た目が画面と完全に一致する。
- **Q. SSE がコネクションを張りっぱなしになるのは？**
  A. `EventHub` の購読者はユーザー単位のチャネル 1 本（バッファ 8）。読み手が遅くて溢れたらそのイベントは捨てる — 通知の取りこぼしより、対戦の記録が止まらないことを優先している（`eventhub.go:13-17`）。購読解除の関数を必ず呼ぶよう `Subscribe` が返す。
- **Q. CSV の上限は？**
  A. `csvExportLimit` で制限している。無制限にすると 1 リクエストでサーバーのメモリを使い切れてしまう（`handler/match.go:197-199`）。

---

# Minor モジュール

## 8. Web — ORM の使用

**種別**: Minor（1pt） / **担当**: kanahash

### 課題要件
データベースに ORM を使う。

### 構成：GORM（モデル定義 + クエリ） + Atlas（マイグレーション）

```
infrastructure/model.go   ← GORM の構造体タグでスキーマを定義（唯一の正本）
        │
        │ go run ./cmd/migrate   （構造体 → DDL を出力）
        ▼
   Atlas が「あるべき姿」として読む
        │
        │ atlas migrate diff <name> --env gorm
        ▼
migrations/*.sql          ← レビュー可能な SQL として生成・コミット
        │
        │ atlas migrate apply（バックエンド起動時に自動実行）
        ▼
   PostgreSQL 16
```

### 該当ファイル

| ファイル | 役割 |
| --- | --- |
| `backend/infrastructure/model.go` | GORM モデル 10 個（228 行）... （残り36 KB）