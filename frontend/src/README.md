# フロントエンドの構成（Feature-Sliced Design）

## レイヤー

上から下へのみ import できる。下から上、および同じレイヤー同士の import は禁止。

| レイヤー                 | 役割                                       | 現状                          |
| ------------------------ | ------------------------------------------ | ----------------------------- |
| `app/`（リポジトリ直下） | Next.js App Router。ルーティングの宣言のみ | `page.tsx` / `login/page.tsx` |
| `views/`                 | 1 画面ぶんの組み立て                       | `home` `login`                |
| `widgets/`               | 複数 feature をまとめた再利用ブロック      | 未使用                        |
| `features/`              | ユーザーの操作単位                         | `auth-google`                 |
| `entities/`              | ビジネス上の実体（user, match など）       | 未使用                        |
| `shared/`                | どこからでも使う土台                       | `ui` `lib` `config`           |

FSD 本来の `pages` レイヤーは Next.js の Pages Router と名前が衝突するため、
公式ガイドに従って `views` に読み替えている。

`widgets/` と `entities/` はまだ中身が無いのでディレクトリを作っていない。
必要になった時点で切ること。空のレイヤーを先に作らないのが FSD の方針。

## スライスとセグメント

各スライスは `ui` / `model` / `api` / `lib` / `config` のセグメントに分け、
`index.ts` で公開するものだけを外に出す。

```
features/auth-google/
  ui/google-login-button.tsx
  model/login-error.ts
  index.ts                     ← 外からはここ経由でのみ import
```

```ts
// ✅
import { GoogleLoginButton } from "@/features/auth-google";
// ❌ 内部実装への直接 import
import { GoogleLoginButton } from "@/features/auth-google/ui/google-login-button";
```

## スタイル

- Tailwind v4 + shadcn/ui。UI プリミティブは `pnpm dlx shadcn@latest add <name>` で
  `shared/ui/` に追加される（出力先は `components.json` の aliases で指定済み）。
- 配色は `app/globals.css` の `.dark` ブロックに集約。dark 基調に統一している。
- **コンポーネント側に `slate-800` や `emerald-500` を直接書かないこと。**
  `bg-card` `text-muted-foreground` `bg-primary` のようなセマンティックトークンを使う。
  例外は Google ボタンの白地のようなブランド指定のみ。
