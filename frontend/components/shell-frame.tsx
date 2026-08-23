"use client";

import { useEffect, type ReactNode } from "react";
import { usePathname } from "next/navigation";
import { Burger, Drawer, ScrollArea } from "@mantine/core";
import { useDisclosure, useMediaQuery } from "@mantine/hooks";
import { useTranslations } from "next-intl";

/** ログイン後の外枠の見た目だけを持つ部品。
 *
 * md 以上はサイドバー常設、md 未満はヘッダーのハンバーガーで開く Drawer に畳む。
 * 開閉状態を持つのでクライアント側だが、中身（brand / nav）は Server Component の
 * ままツリーごと props で受け取るので、ここに描画対象は増えない。
 */
export function ShellFrame({
  brand,
  nav,
  children,
}: {
  brand: ReactNode;
  nav: ReactNode;
  children: ReactNode;
}) {
  const t = useTranslations("nav");
  const pathname = usePathname();
  const [opened, { toggle, close }] = useDisclosure(false);

  // 画面を広げたときに開きっぱなしにしない。Drawer は開いている間 body の
  // スクロールを止めるので、サイドバー常設幅では必ず畳んでおく必要がある。
  const desktop = useMediaQuery("(min-width: 48em)");
  useEffect(() => {
    if (desktop) close();
  }, [desktop, close]);

  // 遷移したら閉じる。メニューから飛ぶたびに手で閉じさせない。
  useEffect(() => {
    close();
  }, [pathname, close]);

  return (
    <div className="flex min-h-dvh flex-col bg-body md:flex-row">
      <header className="sticky top-0 z-20 flex items-center gap-3 border-b border-default-border bg-dark-600 px-4 py-2 md:hidden">
        <Burger
          opened={opened}
          onClick={toggle}
          size="sm"
          aria-label={opened ? t("closeMenu") : t("menu")}
        />
        {brand}
      </header>

      <aside className="hidden shrink-0 flex-col gap-4 border-r border-default-border bg-dark-600 p-4 md:flex md:w-64">
        <div className="px-2 py-3">{brand}</div>
        {nav}
      </aside>

      <Drawer
        opened={opened}
        onClose={close}
        title={brand}
        size="17rem"
        padding="md"
        // 常設サイドバーと同じ面の色に揃える（既定は body 色で一段暗い）。
        classNames={{ content: "bg-dark-600", header: "bg-dark-600" }}
        styles={{
          content: { display: "flex", flexDirection: "column" },
          body: { flex: 1, display: "flex", flexDirection: "column" },
        }}
      >
        <div className="flex flex-1 flex-col gap-4">{nav}</div>
      </Drawer>

      <ScrollArea className="flex-1" scrollbars="y">
        <div className="p-4 sm:p-6 md:p-10">{children}</div>
      </ScrollArea>
    </div>
  );
}
