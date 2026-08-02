# フロントエンド 言語TypeScript

Next.js 16（App Router）/ React 19 / TailwindCSS v4。パッケージマネージャは pnpm。
設計は**Feature-Sliced Design**。

## 起動する

```bash
# Docker（ルートから）
make up                 # http://localhost:3000
make exec-frontend

# ホストで直接
corepack enable pnpm            # package.json でピン留めした pnpm が使えるようになる
pnpm install --frozen-lockfile
pnpm dev                        # :3000
```

前提は Node.js v22.13 以上 / pnpm v11.17.0（`package.json` の `engines` と `packageManager`）。

| コマンド            | 内容                               |
| ------------------- | ---------------------------------- |
| `pnpm dev`          | 開発サーバー（ホットリロード）     |
| `pnpm build`        | 本番ビルド                         |
| `pnpm start`        | 本番サーバー（`build` 済みが前提） |
| `pnpm lint`         | ESLint で静的解析                  |
| `pnpm lint:fix`     | ESLint で自動修正                  |
| `pnpm format`       | Prettier で整形                    |
| `pnpm format:check` | 整形漏れの検出（CI と同じ）        |

> Docker では `./typescript` をマウントしているのでホットリロードが効くが、`node_modules` は匿名ボリュームでマスクされている。**依存を追加したら `docker compose build frontend` でイメージを作り直す。**

## App Router の仕組み

**`app/` 以下のディレクトリ構造がそのまま URL になる。** ルーティング設定ファイルは書かない。

```
app/
├── layout.tsx          →  全ページ共通の外枠（必須）
├── page.tsx            →  /
├── game/
│   └── page.tsx        →  /game
└── users/
    └── [id]/
        └── page.tsx    →  /users/123  （[id] は動的セグメント）
```

ファイル名には決まった役割がある。

| ファイル        | 役割                                                            |
| --------------- | --------------------------------------------------------------- |
| `page.tsx`      | そのパスの画面本体。これがあって初めて URL としてアクセスできる |
| `layout.tsx`    | 配下のページを包む外枠。ページ遷移しても再マウントされない      |
| `loading.tsx`   | データ取得中に自動で表示される                                  |
| `error.tsx`     | エラー時に自動で表示される（Client Component 必須）             |
| `not-found.tsx` | 404 のときに表示される                                          |
| `route.ts`      | 画面ではなく API エンドポイントを作る場合                       |

ルートレイアウト（[app/layout.tsx](app/layout.tsx)）だけは `<html>` と `<body>` を自分で書く。
`metadata` を export するだけで `<head>` が生成され、ページ側で同じものを export すれば上書きできる。
フォントは `next/font` でビルド時に取り込むため、実行時に Google へのリクエストは飛ばない。

## Server Component と Client Component

App Router では **すべてのコンポーネントがデフォルトで Server Component**。
これが Pages Router との一番大きな違い。

|                          | Server Component（既定）     | Client Component（`"use client"`） |
| ------------------------ | ---------------------------- | ---------------------------------- |
| どこで動くか             | サーバーのみ                 | サーバー（初回 HTML）＋ブラウザ    |
| `useState` / `useEffect` | 使えない                     | 使える                             |
| `onClick` などのイベント | 使えない                     | 使える                             |
| DB や秘密鍵へのアクセス  | できる（ブラウザに漏れない） | できない                           |
| JS バンドルサイズ        | 増えない                     | 増える                             |

```tsx
// Server Component（既定）: async にして直接 await できる
export default async function Page() {
  const res = await fetch("http://backend:4000");
  const data = await res.json();
  return <p>{data.message}</p>;
}
```

```tsx
"use client"; // ← ファイル先頭に書くと Client Component になる

import { useState } from "react";

export default function Counter() {
  const [n, setN] = useState(0);
  return <button onClick={() => setN(n + 1)}>{n}</button>;
}
```

使い分けの方針:

1. **まず Server Component で書く**
2. 状態・イベントハンドラ・ブラウザ API（`window`、`localStorage`）が必要になったら、**その部分だけ**を小さな Client Component に切り出す
3. ページ全体に `"use client"` を付けるのは避ける（配下すべてがクライアント側に送られてしまう）

## ディレクトリ構成

```
frontend/
├── app/                  ルーティングと画面（App Router）
│   ├── layout.tsx        ルートレイアウト
│   ├── page.tsx          トップページ
│   └── globals.css       Tailwind の読み込みとテーマ定義
├── types/                DB スキーマに対応する型定義
├── public/               静的ファイル（/next.svg のように参照）
├── eslint.config.mjs     ESLint 設定（Flat Config）
├── .prettierrc.json      Prettier 設定
├── next.config.ts        Next.js 設定
├── postcss.config.mjs    TailwindCSS を PostCSS プラグインとして登録
└── tsconfig.json         TypeScript 設定（strict / パスエイリアス）
```

`tsconfig.json` の `paths` で `@/*` がプロジェクトルートを指すので、`../../` を書かずに済む:

```tsx
import type { User } from "@/types/models";
```

### 設計の方向: Feature-Sliced Design

バックエンドの Clean Architecture と同じく、**依存は一方向にだけ流れる**。
FSD では層に序列があり、**上の層は下の層を import してよいが、逆は禁止**。

```
app        アプリの起動点。ルーティング、テーマ、i18n などの設定まわり
   │
widgets    ユースケース単位の UI。Card・Table・Form を組み合わせた部品の集合体
   │
features   振る舞い単位の処理。ログイン、マッチメイキング、駒を動かす
   │
entities   ドメインモデル。User・Match といった名詞ごとの型とロジック
   │
shared     どこからでも使える汎用部品。UI・hooks・lib・types
```

同じ層の中では**スライス同士を直接 import しない**（`features/login` から `features/matchmaking` を呼ばない）。
共有したくなったら一段下の層へ下ろす。これを守ると、機能を消すときにディレクトリごと消せる。

現状は `app/` と `types/`（将来 `shared/types` に相当）だけ。画面が増えるタイミングで上記の層を切っていく。
層の定義・スライスとセグメントの切り方・Next.js App Router との噛み合わせは [docs/FSD.md](../docs/FSD.md)、
原典は [公式ドキュメント](https://feature-sliced.github.io/documentation/)。

## 型定義とスキーマの同期

`types/models.ts` は DB スキーマの TypeScript 版で、**手作業で同期させる必要がある**。

```
docs/database-design.md          設計の正本（人が読む）
        │  写す
        ├──────────────────────────────┐
        ▼                              ▼
backend/infrastructure/model.go  frontend/types/models.ts
   GORM の構造体（→ Atlas）         フロントの型
```

押さえておくこと:

- **スキーマを変えたら 3 箇所すべてを直す。** 自動生成は無いので、ここだけは仕組みで守られていない
- **`types/models.ts` は「テーブル定義」であって「API のレスポンス形」ではない。** `passwordHash` や `tokenHash` のようにクライアントへ絶対に送ってはいけない項目を含む。API のレスポンス型はこれを直接使わず、必要な項目だけを選んだ別の型を定義する
- `Timestamp` は `string`。Go 側は `time.Time`（`timestamptz`）だが JSON を経由すると文字列になるため、`Date` として扱いたい場合は受け取り側で変換する
- `owner?` `user?` のような省略可能プロパティは **JOIN して取得したときだけ入る**。存在を前提にしない

## TailwindCSS v4

v3 までと違い **`tailwind.config.js` は無い**。設定は CSS に直接書く。

- [postcss.config.mjs](postcss.config.mjs) … `@tailwindcss/postcss` を登録するだけ
- [app/globals.css](app/globals.css) … `@import "tailwindcss";` で本体を読み込み、`@theme inline { ... }` でデザイントークン（色・フォント）を定義

```css
@import "tailwindcss";

@theme inline {
  --color-background: var(--background); /* bg-background が使えるようになる */
  --font-sans: var(--font-geist-sans); /* font-sans が Geist になる */
}
```

`@theme` に定義した変数はそのままユーティリティクラス名になる。
ダークモードは `prefers-color-scheme` で CSS 変数を差し替えつつ、クラス側は `dark:bg-black` のように書く。

## バックエンドを呼ぶ

呼び出し元がブラウザかサーバーかで、使う URL が違う。

| 呼び出し元                       | URL                                                          |
| -------------------------------- | ------------------------------------------------------------ |
| Client Component / ブラウザ      | `process.env.NEXT_PUBLIC_API_URL`（`http://localhost:4000`） |
| Server Component / Route Handler | `http://backend:4000`（コンテナ間通信）                      |

`NEXT_PUBLIC_` が付いた環境変数**だけ**がブラウザ向けのバンドルに埋め込まれる。
逆に言うと、**秘密にしたい値に `NEXT_PUBLIC_` を付けてはいけない**。

許可されるオリジンはバックエンド側の `backend/main.go` の `allowedOrigins` にある。
サービス間の関係はルートの [README.md](../README.md#サービス間の関係) を参照。

## Lint / Format

役割を分けている。

- **Prettier**（[.prettierrc.json](.prettierrc.json)）… 整形だけを担当
- **ESLint**（[eslint.config.mjs](eslint.config.mjs)）… バグや規約違反の検出だけを担当

`eslint-config-prettier` を設定の**最後**に置いて ESLint 側の整形系ルールを無効化しているので、
両者が衝突して直し合いになることはない。

```bash
pnpm format      # 整形する
pnpm lint        # 静的解析
pnpm lint:fix    # 自動修正できるものを直す
```

CI は [frontend-lint-format.yml](../.github/workflows/frontend-lint-format.yml)。
`frontend/**` に変更がある PR / push で走るので、push 前に `pnpm format && pnpm lint` を通しておくと落ちない。

レビュー観点は [pull_request_template.md](../.github/pull_request_template.md) にある。
