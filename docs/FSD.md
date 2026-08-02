# Feature-Sliced Design (FSD)

フロントエンド専用のアーキテクチャ。**ファイルをどこに置くか**を、毎回同じ手順で決められるようにするのが目的。

- 原典: [公式ドキュメント](https://feature-sliced.github.io/documentation/)
- 解説記事: [zenn / learn-fsd-design](https://zenn.dev/nyaomaru/articles/learn-fsd-design)

## 何が嬉しいのか

`components/` `hooks/` `utils/` のように**技術で切る**と、機能が増えるほど 1 機能のコードが全ディレクトリに散らばる。
FSD は**機能で切って、そのうえで依存の向きを固定する**。結果としてこうなる:

- 機能を消すとき、**ディレクトリごと消せる**（他所から呼ばれていないことがルールで保証される）
- 「この修正はどこまで影響するか」が、依存の向きから機械的にわかる
- 新しく入った人でも、置き場所で迷わない

バックエンドの Clean Architecture と発想は同じ。**依存が一方向にしか流れない**ことに全部を賭けている。

## 3 つの軸

FSD は 3 段階で場所を決める。

```
レイヤー（縦の序列）      app / views / widgets / features / entities / shared
   └ スライス（横の区切り） features/login, entities/user  … ドメイン単位
       └ セグメント（中身）  ui/ model/ api/ lib/          … 技術的な役割
```

```
features/login/          ← スライス（features レイヤーの中の1つ）
├── ui/LoginForm.tsx     ← セグメント
├── model/useLogin.ts
├── api/login.ts
└── index.ts             ← Public API
```

---

## 1. レイヤー — 縦の序列

**上の層は下の層を import してよい。下から上は絶対に禁止。**

| 層          | 何を置くか                                                 | 例                                         |
| ----------- | ---------------------------------------------------------- | ------------------------------------------ |
| `app/`      | 起動点。ルーティング、共通レイアウト、Provider、全体の CSS | `layout.tsx`、`globals.css`                |
| `views/`    | 1 画面ぶんの組み立て（FSD 本来の `pages` 層）              | 対戦画面、トーナメント画面                 |
| `widgets/`  | 画面の一区画として自立した UI ブロック                     | ヘッダー、チャットパネル、対戦履歴テーブル |
| `features/` | ユーザーが起こす「動詞」。1 機能 = 1 スライス              | ログイン、フレンド申請、マッチ参加         |
| `entities/` | ドメインの「名詞」。型・表示部品・取得処理                 | User、Match、Tournament                    |
| `shared/`   | ドメインを一切知らない汎用部品                             | Button、fetch ラッパー、hooks、types       |

```
app        ← 一番外側。下を全部使える
  ↓
views      ← widgets 以下を使える
  ↓
widgets    ← features 以下を使える
  ↓
features   ← entities, shared を使える
  ↓
entities   ← shared だけ使える
  ↓
shared     ← 何にも依存しない（一番内側）
```

### 層の見分け方

判断に迷ったら「**この層を丸ごと消したとき、下の層は壊れないか？**」を確認する。壊れるなら層を間違えている。

- `shared/Button` は「ログイン」を知らない → だから `shared`
- `entities/user` は「User をどう表示するか」は知っているが、「ログインする」は知らない → だから `entities`
- `features/login` は User を使ってログイン処理をする → だから `features`
- `widgets/header` はログインボタンとユーザー名を並べる → だから `widgets`

**`features` と `entities` の切り分けが一番迷う。動詞なら `features`、名詞なら `entities`** と覚える。
「フレンド申請する（動詞）」は `features/add-friend`、「User の型と表示（名詞）」は `entities/user`。

### `shared/` に何を入れるか

**ドメインを知らないもの**だけ。`shared/ui/UserCard.tsx` は間違い（User を知っているので `entities/user/ui/` へ）。

```
shared/
├── ui/       Button, Input, Modal … 汎用コンポーネント
├── api/      fetch ラッパー、共通エラーハンドリング
├── lib/      日付整形、クラス名結合などのユーティリティ
├── config/   環境変数、定数
└── types/    汎用型（このプロジェクトでは既存の types/ が該当）
```

---

## 2. スライス — 層を横に切る

各層の中を**ドメイン単位**で区切る。`features/login`、`entities/user` がそれぞれ 1 スライス。

### 同じ層のスライス同士は import しない

これが FSD で**一番よく破られ、一番重要なルール**。

```ts
// ❌ features/matchmaking から features/login を import
import { useLogin } from "@/features/login";

// ✅ 共有したいものは一段下の層へ下ろす
import { useCurrentUser } from "@/entities/user";
```

共有したくなったら**下の層へ下ろす**。両方が使うドメインの知識なら `entities/`、ドメインを知らない部品なら `shared/`。

このルールがあるおかげで、`features/login` を消しても他の feature は壊れないことが**確認しなくてもわかる**。

### スライスが無い層

`app/` と `shared/` にはスライスが無い。ドメインで切る意味がないため、直接セグメント（`shared/ui/` など）から始まる。

### 例外: `@x` 記法

`entities` 同士がどうしても参照し合う場合（`Match` が `User` を持つなど）だけ、**明示的な cross-import** が公式に認められている。

```
entities/user/
└── @x/match.ts      ← match スライス専用に公開する入口
```

```ts
// entities/match/model/types.ts
import type { User } from "@/entities/user/@x/match";
```

「例外的にここだけ繋がっている」ことがパスから見えるのが狙い。**多用したら設計を疑う**。

---

## 3. セグメント — スライスの中身

スライスの中は**技術的な役割**で切る。

| セグメント | 中身                                                       |
| ---------- | ---------------------------------------------------------- |
| `ui/`      | 見た目。コンポーネント                                     |
| `model/`   | 状態とロジック。store、スキーマ、型、変換                  |
| `api/`     | サーバーとの通信                                           |
| `lib/`     | そのスライス専用のユーティリティ                           |
| `config/`  | そのスライス専用の定数                                     |
| `index.ts` | **Public API**。外に見せるものだけを re-export（後述）     |

全部作る必要はない。型しか無いスライスなら `model/` と `index.ts` だけでいい。

### Public API — `index.ts` 越しにしか import しない

```ts
// features/login/index.ts
export { LoginForm } from "./ui/LoginForm";
export { useLogin } from "./model/useLogin";
// api/login.ts は export しない = 外から呼べない
```

```ts
// ✅ Public API 経由
import { LoginForm } from "@/features/login";

// ❌ 内部に直接手を伸ばしている
import { LoginForm } from "@/features/login/ui/LoginForm";
```

これを守ると、**スライスの内部構造をいくら変えても外側を直さなくて済む**。
逆に守らないと、内部ファイルを 1 つ動かすだけで各所が壊れ、FSD の利点がほぼ全部消える。

---

## Next.js App Router との噛み合わせ

Next.js では `app/` が URL と直結した**予約ディレクトリ**なので、FSD をそのままの名前では適用できない。

| FSD の層  | Next.js での扱い                                                 |
| --------- | ---------------------------------------------------------------- |
| `app/`    | Next.js の `app/`。ただし**ルーティングと共通設定だけ**に留める   |
| `pages/`  | 名前が Pages Router と衝突するため **`views/` にリネーム**して使う |
| それ以降  | そのまま（`widgets/` `features/` `entities/` `shared/`）          |

`app/**/page.tsx` は **`views/` を呼ぶだけの薄い入口**にする。

```tsx
// app/tournament/[id]/page.tsx — ルーティングだけを担当する
import { TournamentView } from "@/views/tournament";

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <TournamentView id={id} />;
}
```

こうする理由:

- 画面のロジックが Next.js の規約から切り離され、**ルーティングを変えても画面は動く**
- `page.tsx` が肥大化しない
- `"use client"` の境界を `views/` 側の必要な部分だけに閉じ込められる

---

## 置き場所の決め方（フローチャート）

```
そのコードは…

ドメイン（User, Match…）を知らない？
    └─ YES → shared/

特定の名詞の型・表示・取得？
    └─ YES → entities/<名詞>/

ユーザーが起こす動作（ログインする、申請する）？
    └─ YES → features/<動詞>/

複数の feature/entity を並べた、画面の一区画？
    └─ YES → widgets/<区画名>/

1 画面まるごとの組み立て？
    └─ YES → views/<画面名>/

ルーティング・Provider・全体設定？
    └─ YES → app/
```

## よくある間違い

| 間違い                                             | どうするか                                       |
| -------------------------------------------------- | ------------------------------------------------ |
| `shared/ui/UserCard.tsx`                           | User を知っているので `entities/user/ui/` へ      |
| `features/a` から `features/b` を import           | 共通部分を `entities/` か `shared/` へ下ろす      |
| `@/features/login/ui/LoginForm` と深く import      | `index.ts` に export して `@/features/login` へ   |
| `entities/` に fetch もフォームも全部入れる        | 動作は `features/` へ切り出す                     |
| 最初から 6 層すべてのディレクトリを空で作る        | 必要になった層から切る                            |
| `app/page.tsx` に画面ロジックを直接書く            | `views/` に出して `app/` は入口だけにする         |

## このプロジェクトの現状と導入順

現状は `app/` と `types/`（将来の `shared/types`）だけ。**層を最初から全部作る必要はない。**

1. 画面が 2 つ以上になったら `views/` を切って `app/` を薄くする
2. `types/models.ts` を使う UI が出てきたら `entities/<名詞>/` へ移す
3. 同じボタンや fetch を 2 回書いたら `shared/` へ下ろす
4. ログイン・マッチメイキングなど動作が増えたら `features/` を切る

フロント側のセットアップやコマンドは [frontend/README.md](../frontend/README.md) を参照。
