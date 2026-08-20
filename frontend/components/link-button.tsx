"use client";

import type { ReactNode } from "react";
import Link, { type LinkProps } from "next/link";
import { Button, type ButtonProps } from "@mantine/core";

type LinkButtonProps = ButtonProps & LinkProps & { children?: ReactNode };

/** Mantine の Button を next/link として描画する。
 * `component={Link}` を Server Component から直接渡すと
 * 「Functions cannot be passed directly to Client Components」で落ちるため、
 * 結線はこの Client Component の中で閉じる。
 */
export function LinkButton({ children, ...props }: LinkButtonProps) {
  return (
    <Button component={Link} {...props}>
      {children}
    </Button>
  );
}
