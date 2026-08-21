import { cookies } from "next/headers";

import { serverApiUrl } from "@/lib/api";

export interface StatsSummary {
  wins: number;
  losses: number;
  draws: number;
  total_matches: number;
  win_rate: number;
  current_streak: number;
  best_streak: number;
  rating: number;
  ranking: number;
  total_players: number;
  level: number;
  xp: number;
  xp_for_level: number;
  xp_for_next_level: number;
}

export interface TimeseriesPoint {
  date: string;
  wins: number;
  losses: number;
  draws: number;
  matches: number;
  rating: number;
}

export interface Timeseries {
  interval: "day" | "week";
  points: TimeseriesPoint[];
}

export interface BreakdownSlice {
  key: string;
  count: number;
}

export interface Breakdown {
  by_result_type: BreakdownSlice[];
  by_mode: BreakdownSlice[];
  by_outcome: BreakdownSlice[];
}

export interface MatchUser {
  id: string;
  display_name: string;
  handle: string;
  level: number;
}

export interface Match {
  id: string;
  mode: string;
  opponent: MatchUser;
  outcome: "win" | "loss" | "draw";
  result_type: string;
  rating_before: number;
  rating_after: number;
  rating_diff: number;
  xp_gained: number;
  total_moves: number;
  started_at: string;
  finished_at: string;
}

export interface MatchList {
  items: Match[];
  total: number;
  limit: number;
  offset: number;
}

export interface LeaderboardEntry {
  rank: number;
  user: MatchUser;
  rating: number;
  wins: number;
  losses: number;
  win_rate: number;
}

export interface Leaderboard {
  items: LeaderboardEntry[];
  me: LeaderboardEntry | null;
  total: number;
  limit: number;
  offset: number;
}

export interface Achievement {
  code: string;
  name: string;
  description: string;
  category: string;
  progress: number;
  target: number;
  unlocked: boolean;
  unlocked_at: string | null;
}

export interface AchievementList {
  items: Achievement[];
  unlocked_count: number;
  total_count: number;
}

/** 絞り込み。URL のクエリをそのまま backend へ渡すための形。 */
export interface StatsFilters {
  from?: string;
  to?: string;
  mode?: string;
  outcome?: string;
  interval?: string;
  limit?: string;
  offset?: string;
}

/** searchParams から、backend が知っているキーだけを拾う。
 * 知らないキーを素通しすると、URL を触っただけで 400 を返しうるため。
 */
export function toQuery(filters: StatsFilters): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value) params.set(key, value);
  }

  return params.toString();
}

/** Server Component から backend を叩く。Cookie は自動で乗らないので転送する。 */
async function fetchJSON<T>(path: string): Promise<T> {
  const cookie = (await cookies()).toString();

  const res = await fetch(serverApiUrl(path), {
    headers: { cookie },
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`${path} が ${res.status} を返しました`);
  }

  return res.json();
}

export function getSummary(filters: StatsFilters) {
  return fetchJSON<StatsSummary>(`/stats/me?${toQuery(filters)}`);
}

export function getTimeseries(filters: StatsFilters) {
  return fetchJSON<Timeseries>(`/stats/me/timeseries?${toQuery(filters)}`);
}

export function getBreakdown(filters: StatsFilters) {
  return fetchJSON<Breakdown>(`/stats/me/breakdown?${toQuery(filters)}`);
}

export function getMatches(filters: StatsFilters) {
  return fetchJSON<MatchList>(`/matches?${toQuery(filters)}`);
}

export function getLeaderboard(filters: StatsFilters) {
  return fetchJSON<Leaderboard>(`/leaderboard?${toQuery(filters)}`);
}

export function getAchievements() {
  return fetchJSON<AchievementList>("/achievements/me");
}
