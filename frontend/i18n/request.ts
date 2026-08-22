import { cookies } from "next/headers";
import { getRequestConfig } from "next-intl/server";

import { defaultLocale, isLocale } from "@/i18n/config";

/** URL にロケールを載せないルーティング無し構成。
 * Cookie で決め、無ければ日本語。ログイン中の恒久設定（users.preferred_locale）は
 * 切り替え時に一緒に保存され、次回以降も Cookie 経由で反映される。
 */
export default getRequestConfig(async () => {
  const store = await cookies();
  const cookieLocale = store.get("NEXT_LOCALE")?.value;
  const locale = isLocale(cookieLocale) ? cookieLocale : defaultLocale;

  return {
    locale,
    messages: (await import(`../messages/${locale}.json`)).default,
  };
});
