"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { Anchor } from "@mantine/core";

/** 認証画面のページ間リンク。Server Component からは Mantine の Anchor に
 * component={Link}（関数）を渡せないため、ここだけ Client Component にする。
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
