"use client";

import { useSearchParams } from "next/navigation";
import { Button, Group } from "@mantine/core";
import { IconFileTypeCsv, IconPrinter } from "@tabler/icons-react";

import { apiUrl } from "@/lib/api";

/** CSV のダウンロードと、ブラウザの印刷による PDF 出力。
 *
 * PDF をサーバーで作らないのは、チャートの見た目をそのまま出したいため。
 * 印刷用の CSS でサイドバーと操作ボタンを隠してある（globals.css）。
 */
export function ExportButtons() {
  const searchParams = useSearchParams();

  // 画面で絞った条件をそのまま CSV にも効かせる。
  // ページングは付けない（サーバー側が全件を返す）。
  const params = new URLSearchParams();
  for (const key of ["from", "to", "mode", "outcome", "opponent"]) {
    const value = searchParams.get(key);
    if (value) params.set(key, value);
  }

  return (
    <Group gap="xs" className="print:hidden">
      <Button
        component="a"
        href={apiUrl(`/matches/export.csv?${params.toString()}`)}
        variant="default"
        size="xs"
        leftSection={<IconFileTypeCsv size={16} />}
      >
        CSV
      </Button>
      <Button
        onClick={() => window.print()}
        variant="default"
        size="xs"
        leftSection={<IconPrinter size={16} />}
      >
        PDF
      </Button>
    </Group>
  );
}
