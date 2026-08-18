import { LoginPage } from "@/views/login";

// searchParams の取り出しだけこの層で行い、画面本体には素の値を渡す。
// こうしておくと LoginPage 側は Next のルーティング事情を知らずに済む。
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;

  return <LoginPage error={error} />;
}
