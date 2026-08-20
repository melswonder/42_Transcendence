"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { NavLink } from "@mantine/core";
import { IconActivity, IconSettings, IconTrophy } from "@tabler/icons-react";

// 画面がまだ無いものは disabled にしておく。作った時点で外す。
const items = [
  { href: "/", label: "プレイ", icon: IconActivity },
  {
    href: "/leaderboard",
    label: "ランキング",
    icon: IconTrophy,
    disabled: true,
  },
  { href: "/settings", label: "設定", icon: IconSettings, disabled: true },
];

/** 現在地の判定に usePathname を使うのでクライアント側に置く。 */
export function SideNav() {
  const pathname = usePathname();

  return (
    <>
      {items.map(({ href, label, icon: Icon, disabled }) =>
        disabled ? (
          <NavLink
            key={href}
            component="button"
            label={label}
            leftSection={<Icon size={20} />}
            variant="light"
            disabled
          />
        ) : (
          <NavLink
            key={href}
            component={Link}
            href={href}
            label={label}
            leftSection={<Icon size={20} />}
            active={pathname === href}
            variant="light"
          />
        ),
      )}
    </>
  );
}
