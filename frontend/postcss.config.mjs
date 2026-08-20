// Mantine と Tailwind を併用する。プラグインの順序は
// 「Mantine の関数を展開してから Tailwind に渡す」ため、この並びを崩さないこと。
//
// - postcss-preset-mantine : light-dark() / rem() / em() などの Mantine 専用関数を展開する
// - postcss-simple-vars    : CSS モジュール内で $mantine-breakpoint-* を使えるようにする
//   （値は Mantine のデフォルトブレークポイントと必ず一致させること）
// - @tailwindcss/postcss   : Tailwind v4 本体
const config = {
  plugins: {
    "postcss-preset-mantine": {},
    "postcss-simple-vars": {
      variables: {
        "mantine-breakpoint-xs": "36em",
        "mantine-breakpoint-sm": "48em",
        "mantine-breakpoint-md": "62em",
        "mantine-breakpoint-lg": "75em",
        "mantine-breakpoint-xl": "88em",
      },
    },
    "@tailwindcss/postcss": {},
  },
};

export default config;
