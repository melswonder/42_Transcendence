/** ロケールの定義。クライアント・サーバー両方から import できる純粋な定数。 */
export const locales = ["ja", "en", "fr"] as const;
export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "ja";

/** ロケールを覚えておく Cookie。リロードしても言語が保たれる。 */
export const localeCookie = "NEXT_LOCALE";

export function isLocale(v: string | undefined): v is Locale {
  return locales.includes(v as Locale);
}
