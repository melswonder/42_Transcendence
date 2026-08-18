import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

// 変数名は globals.css の @theme inline が参照している --font-sans / --font-mono に合わせる。
// ここを変えると Tailwind の font-sans / font-mono が解決できなくなるので注意。
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
  description: "ブラウザで対戦できる Pong",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    // dark 基調に統一しているため、テーマクラスは固定で付ける。
    // shadcn のコンポーネントが持つ dark: バリアントもこれで有効になる。
    <html lang="ja" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
