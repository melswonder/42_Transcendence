// Mantine の rem() / light-dark() を展開してから Tailwind に渡すので、この並びは崩さない。
// breakpoint の値は Mantine のデフォルトと一致させること。
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
