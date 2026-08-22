"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { NavLink } from "@mantine/core";
import {
  IconActivity,
  IconAward,
  IconChartBar,
  IconEye,
  IconHistory,
  IconSettings,
  IconTrophy,
  IconUsers,
} from "@tabler/icons-react";

const items = [
  { href: "/", label: "プレイ", icon: IconActivity },
  { href: "/watch", label: "観戦", icon: IconEye },
  { href: "/stats", label: "統計", icon: IconChartBar },
  { href: "/matches", label: "対戦履歴", icon: IconHistory },
  { href: "/achievements", label: "実績", icon: IconAward },
  { href: "/leaderboard", label: "ランキング", icon: IconTrophy },
  { href: "/friends", label: "フレンド", icon: IconUsers },
  { href: "/settings", label: "設定", icon: IconSettings },
];

/** 現在地の判定に usePathname を使うのでクライアント側に置く。 */
export function SideNav() {
  const pathname = usePathname();

  return (
    <>
      {items.map(({ href, label, icon: Icon }) => (
        <NavLink
          key={href}
          component={Link}
          href={href}
          label={label}
          leftSection={<Icon size={20} />}
          active={pathname === href}
          variant="light"
        />
      ))}
    </>
  );
}
