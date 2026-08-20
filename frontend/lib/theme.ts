// Button.extend() は Mantine のコンポーネント（"use client" 付き）の静的メソッドで、
// Server Component からは呼べない（クライアント参照になり関数が生えていない）。
// そのためテーマ定義ごとクライアント側に置く。
"use client";

import {
  Button,
  createTheme,
  Paper,
  type MantineColorsTuple,
} from "@mantine/core";

/** 配色はここ一箇所に集約する。コンポーネント側に #0f172a のような
 * 生の色を書かないこと（Mantine の CSS 変数か theme トークンを使う）。
 *
 * Mantine の色は必ず 10 段（0 が最も明るく 9 が最も暗い）。
 * dark の各段には Mantine 側で決まった役割があるので、対応を崩さないこと。
 *   dark[0] 本文の文字 / dark[3] 補助テキスト / dark[4] ボーダー
 *   dark[6] サーフェス（Paper・入力欄）/ dark[7] ページ背景
 */
const dark: MantineColorsTuple = [
  "#f8fafc",
  "#e2e8f0",
  "#cbd5e1",
  "#94a3b8",
  "#334155",
  "#1e293b",
  "#0f172a",
  "#020617",
  "#010409",
  "#000206",
];

/** アクセント。盤面の駒や壁にも使うので、彩度の高い緑を 1 色だけ持つ。 */
const emerald: MantineColorsTuple = [
  "#ecfdf5",
  "#d1fae5",
  "#a7f3d0",
  "#6ee7b7",
  "#34d399",
  "#10b981",
  "#059669",
  "#047857",
  "#065f46",
  "#064e3b",
];

export const theme = createTheme({
  colors: { dark, emerald },
  primaryColor: "emerald",
  // dark 基調では既定の 8 段目だと沈むので、明るい 5 段目を主役にする。
  primaryShade: { light: 6, dark: 5 },
  // 文字色を背景の明度から自動で決めさせる。emerald の上に白文字が乗って
  // 読めなくなる事故を防ぐ。
  autoContrast: true,
  defaultRadius: "md",
  // next/font が生やす CSS 変数を参照する。変数名は app/layout.tsx と対で管理。
  fontFamily: "var(--font-sans), sans-serif",
  fontFamilyMonospace: "var(--font-mono), monospace",
  headings: {
    fontFamily: "var(--font-sans), sans-serif",
    fontWeight: "700",
  },
  components: {
    // 呼び出し側で size / radius を書き散らさないよう、既定値をここで揃える。
    Button: Button.extend({
      defaultProps: { size: "md", radius: "md" },
    }),
    Paper: Paper.extend({
      defaultProps: { radius: "lg", withBorder: true },
    }),
  },
});
