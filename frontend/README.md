# フロントエンド 言語TypeScript

Next.js 16（App Router）/ React 19 / TailwindCSS v4 + Mantine v9。パッケージマネージャは pnpm。

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

> Docker では `./frontend` をマウントしているのでホットリロードが効くが、`node_modules` は匿名ボリュームでマスクされている。**依存を追加したら、イメージを作り直したうえで匿名ボリュームも捨てる。**
>
> ```bash
> docker compose build frontend
> docker compose up -d -V frontend   # -V を付けないと古い node_modules が残る
> ```
>
> `-V` を忘れると、コンテナ内の pnpm が食い違いに気付いて再インストールを試み、
> `ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY` で起動に失敗する。

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

**`app/` `components/` `lib/` の 3 つだけ。** 迷ったら「URL になるか / 複数の画面から使うか」で決まる。

```
frontend/
├── app/                  ルーティングと画面（App Router）
│   ├── layout.tsx        ルートレイアウト
│   ├── globals.css       Tailwind と Mantine の読み込み・レイヤー順・トークン橋渡し
│   ├── page.tsx          /
│   └── login/page.tsx    /login
├── components/           複数の画面から使う UI 部品
│   ├── google-login-button.tsx
│   ├── link-button.tsx
│   └── quoridor-mark.tsx
├── lib/                  UI ではない共通ロジック
│   ├── theme.ts          Mantine のテーマ（配色・フォント・既定値）
│   ├── api.ts            バックエンドの URL 組み立て
│   └── login-error.ts    エラーコード → 表示文言
├── public/               静的ファイル（/next.svg のように参照）
├── eslint.config.mjs     ESLint 設定（Flat Config）
├── .prettierrc.json      Prettier 設定
├── next.config.ts        Next.js 設定
├── postcss.config.mjs    Mantine プリセットと TailwindCSS を登録
└── tsconfig.json         TypeScript 設定（strict / パスエイリアス）
```

置き場所の決め方:

| 書きたいもの                | 置き場所                             |
| --------------------------- | ------------------------------------ |
| 画面そのもの（URL を持つ）  | `app/<パス>/page.tsx` に**直接書く** |
| その画面でしか使わない部品  | 同じ `page.tsx` の中、または隣に置く |
| 2 つ以上の画面から使う部品  | `components/`                        |
| UI を持たない処理・定数・型 | `lib/`                               |

- **画面を `app/` の外に出さない。** `page.tsx` は「呼ぶだけの薄い入口」にせず、画面の中身をそこに書く。
  どこに何が描かれているかを URL から 1 ホップで辿れる状態を保つため
- **`components/` は最初から入れない。** 2 つ目の画面で使うことになった時点で移す。
  先回りして共通化すると、実際には共通でなかった部品が溜まる
- ファイル名は kebab-case、export は名前付き（`page.tsx` の default export だけ例外）
- 機能が増えて `components/` が平らなまま辛くなったら、その時に
  `components/game/` のようなサブディレクトリで切る。空のディレクトリを先に作らない

`tsconfig.json` の `paths` で `@/*` がプロジェクトルートを指すので、`../../` を書かずに済む:

```tsx
import { apiUrl } from "@/lib/api";
import { QuoridorMark } from "@/components/quoridor-mark";
```

## 型定義とスキーマの同期

DB スキーマの TypeScript 版は `lib/types.ts` に置き、**手作業で同期させる必要がある**
（フロントから DB を触るようになった時点で作る。現状は未作成）。

```
docs/database-design.md          設計の正本（人が読む）
        │  写す
        ├──────────────────────────────┐
        ▼                              ▼
backend/infrastructure/model.go  frontend/lib/types.ts
   GORM の構造体（→ Atlas）         フロントの型
```

押さえておくこと:

- **スキーマを変えたら 3 箇所すべてを直す。** 自動生成は無いので、ここだけは仕組みで守られていない
- **これは「テーブル定義」であって「API のレスポンス形」ではない。** `passwordHash` や `tokenHash` のようにクライアントへ絶対に送ってはいけない項目を含む。API のレスポンス型はこれを直接使わず、必要な項目だけを選んだ別の型を定義する
- `Timestamp` は `string`。Go 側は `time.Time`（`timestamptz`）だが JSON を経由すると文字列になるため、`Date` として扱いたい場合は受け取り側で変換する
- `owner?` `user?` のような省略可能プロパティは **JOIN して取得したときだけ入る**。存在を前提にしない

## TailwindCSS v4 + Mantine v9

**スタイリングの土台は Tailwind、コンポーネントは Mantine** の二段構え。
ボタン・カード・アラート・モーダルのようなプリミティブは自作せず
[@mantine/core](https://mantine.dev/) から import し、既製のブロックは
[ui.mantine.dev](https://ui.mantine.dev/) からコピーして `app/` や `components/` に置く。
そこへの微調整（余白・レイアウト・字送り）を Tailwind のユーティリティで当てる。

### レイヤー順がすべて

同居の要は [app/globals.css](app/globals.css) 冒頭の 1 行。
**後ろのレイヤーほど強い**ので、この順序が優先順位の正本になる。

```css
@layer theme, base, mantine, components, utilities;

@import "tailwindcss/theme.css" layer(theme);
@import "tailwindcss/preflight.css" layer(base);
@import "@mantine/core/styles.layer.css"; /* 中身が @layer mantine で包まれている */
@import "tailwindcss/utilities.css" layer(utilities);
```

- `base`（Tailwind の reset）より `mantine` が後ろ → preflight が Mantine の見た目を壊さない
- `mantine` より `utilities` が後ろ → **`<Paper className="p-0">` のように Tailwind で上書きできる。`!important` は不要**
- `@mantine/core/styles.css` ではなく **`styles.layer.css`** を読むこと。前者はレイヤーに入っておらず、順序制御が効かない
- `@import "tailwindcss"` の一括読み込みも使わない。Mantine を base と utilities の**間**に挟めなくなる

### 色とトークン

配色の実体は [lib/theme.ts](lib/theme.ts)（`createTheme`）にしかない。
Tailwind 側へは `globals.css` の `@theme inline` で橋渡ししてあるので、
**Tailwind のクラスからも同じ色を引ける**。

```tsx
<main className="bg-body">        {/* → var(--mantine-color-body) */}
<p className="text-dimmed">       {/* → var(--mantine-color-dimmed) */}
<span className="bg-emerald-500"> {/* → var(--mantine-color-emerald-5) */}
```

- **生の色を書かないこと。** `bg-[#0f172a]` や `text-slate-400` は禁止。
  配色を変えたい時に `theme.ts` 一箇所で終わらせるため
- 色を足したい時は `theme.ts` に 10 段のタプルで追加し、`globals.css` の
  `@theme inline` に対応行を足す
- 配色は dark 基調に固定（`layout.tsx` の `forceColorScheme="dark"`）。
  切り替えを入れるならそれを外して `useMantineColorScheme` を使う

### どちらで書くかの目安

| やりたいこと                              | 使うもの                                   |
| ----------------------------------------- | ------------------------------------------ |
| ボタン・入力・カード・モーダル・通知      | Mantine のコンポーネント                   |
| 余白・整列・グリッド・レスポンシブ        | Tailwind のユーティリティ                  |
| Mantine コンポーネントの props で足りる分 | Mantine の props（`gap` `c` `maw` `size`） |
| props にも Tailwind にも無い複雑な指定    | 隣に置く `*.module.css`（最後の手段）      |

`*.module.css` では `rem()` と `$mantine-breakpoint-*` が使える
（[postcss.config.mjs](postcss.config.mjs) の `postcss-preset-mantine` /
`postcss-simple-vars`）。素の `px` は書かず `rem()` を通すこと。

### アイコン

Mantine が公式に前提としている [@tabler/icons-react](https://tabler.io/icons) を使う。

```tsx
import { IconArrowRight } from "@tabler/icons-react";

<Button rightSection={<IconArrowRight size={18} />}>始める</Button>;
```

プロダクト固有の意匠（ロゴ、Google の G マーク）だけ `components/*-mark.tsx` に置く。

### Server Component との噛み合わせ

Mantine のコンポーネントは `"use client"` 付きで配布されているので、
**Server Component からそのまま import してよい**。ただし RSC 境界を関数は越えられないため、
次の 2 つは Client Component 側に閉じ込める必要がある。

- `lib/theme.ts` … `Button.extend()` を使うので先頭に `"use client"` が必要
- `component={Link}` … 関数を props で渡すことになるので不可。
  [components/link-button.tsx](components/link-button.tsx) のような小さなラッパーを挟む

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
