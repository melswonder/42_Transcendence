"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { Anchor } from "@mantine/core";

/** 認証画面のページ間リンク。
 * Server Component から Mantine の Anchor に component={Link}（関数）は
 * 渡せないので、ここだけ Client Component に切り出す。
 */
export function AuthLink({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) {
  return (
    <Anchor component={Link} href={href} size="sm">
      {children}
    </Anchor>
  );
}
