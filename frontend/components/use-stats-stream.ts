"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { apiUrl } from "@/lib/api";

/** 対戦が記録されたら画面を取り直す。
 *
 * イベントには統計そのものが載っていない（載せると「イベントで運ばれた値」と
 * 「API で取り直した値」の 2 経路ができて食い違う）。合図として使い、
 * router.refresh() で Server Component 側にもう一度引かせる。
 */
export function useStatsStream() {
  const router = useRouter();

  useEffect(() => {
    // withCredentials を付けないとセッション Cookie が乗らず 401 になる。
    const source = new EventSource(apiUrl("/stats/stream"), {
      withCredentials: true,
    });

    source.addEventListener("match_recorded", () => router.refresh());

    // 切断時の再接続は EventSource 側が自動で行うので、ここでは何もしない。
    // 開発中は握り潰したことに気付けるようログだけ出す。本番で出すと
    // 「コンソールに警告が無いこと」を確かめる邪魔になるので dev 限定。
    source.onerror = () => {
      if (process.env.NODE_ENV !== "production") {
        console.warn("stats stream: connection lost, reconnecting");
      }
    };

    return () => source.close();
  }, [router]);
}

/** 画面に何も足さず、購読だけを行うコンポーネント。
 * Server Component のページからは、これを 1 つ置けば繋がる。
 */
export function StatsStream() {
  useStatsStream();

  return null;
}
