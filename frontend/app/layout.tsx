import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import {
  ColorSchemeScript,
  MantineProvider,
  mantineHtmlProps,
} from "@mantine/core";

// Mantine と Tailwind の読み込みは globals.css に集約している。
// レイヤー順（＝優先順位）をあちこちに散らさないため、ここでは触らない。
import "./globals.css";

import { theme } from "@/lib/theme";

// 変数名は lib/theme.ts の fontFamily が参照している
// --font-sans / --font-mono に合わせる。片方だけ変えると解決できなくなる。
const geistSans = Geist({
  variable: "--font-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Transcendence",
  description: "ブラウザで対戦できるコリドール（Quoridor）",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    // dark 基調に統一しているので配色は固定する。切り替えを許すなら
    // forceColorScheme を外して defaultColorScheme と useMantineColorScheme を使う。
    <html lang="ja" {...mantineHtmlProps} data-mantine-color-scheme="dark">
      <head>
        <ColorSchemeScript forceColorScheme="dark" />
      </head>
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        <MantineProvider theme={theme} forceColorScheme="dark">
          {children}
        </MantineProvider>
      </body>
    </html>
  );
}
