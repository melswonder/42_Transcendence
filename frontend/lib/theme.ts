// Button.extend() が Server Component から呼べないため、テーマごとクライアントに置く。
"use client";

import {
  Button,
  createTheme,
  Paper,
  type MantineColorsTuple,
} from "@mantine/core";

/** 各段には Mantine 側で決まった役割があるので対応を崩さないこと。
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
  // 既定の 8 段目は dark 基調だと沈むので 5 段目を主役にする。
  primaryShade: { light: 6, dark: 5 },
  // emerald の上に白文字が乗って読めなくなる事故を防ぐ。
  autoContrast: true,
  defaultRadius: "md",
  // 変数名は app/layout.tsx と対で管理。
  fontFamily: "var(--font-sans), sans-serif",
  fontFamilyMonospace: "var(--font-mono), monospace",
  headings: {
    fontFamily: "var(--font-sans), sans-serif",
    fontWeight: "700",
  },
  components: {
    Button: Button.extend({
      defaultProps: { size: "md", radius: "md" },
    }),
    Paper: Paper.extend({
      defaultProps: { radius: "lg", withBorder: true },
    }),
  },
});
