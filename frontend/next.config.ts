import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./i18n/request.ts");

const nextConfig: NextConfig = {
  // 別 PC から（Caddy 経由で）dev サーバーを見るときのホスト。
  // 未設定なら localhost だけで、従来と同じ挙動になる。
  allowedDevOrigins: process.env.PUBLIC_HOST ? [process.env.PUBLIC_HOST] : [],
};

export default withNextIntl(nextConfig);
