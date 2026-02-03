import { Container, Title, Text, Stack, Paper } from "@mantine/core";

export function PrivacyWidget() {
  return (
    <Container size="md" my={40}>
      <Title order={1} mb={30}>
        プライバシーポリシー
      </Title>

      <Stack gap="xl">
        <Paper shadow="xs" p="md" withBorder>
          <Text size="sm" c="dimmed" mb="md">
            最終更新日: 2026年2月3日
          </Text>
          <Text>
            本プライバシーポリシーは、チェスVS将棋ゲーム（以下「本サービス」）における、ユーザーの個人情報の取扱いについて定めたものです。
          </Text>
        </Paper>

        <div>
          <Title order={2} size="h3" mb="md">
            1. 収集する情報
          </Title>
          <Text mb="sm">本サービスでは、以下の情報を収集する場合があります：</Text>
          <Stack gap="xs" pl="md">
            <Text>• Googleアカウント情報（氏名、メールアドレス、プロフィール画像）</Text>
            <Text>• ゲームプレイに関する情報（対戦履歴、勝敗記録、レーティング）</Text>
            <Text>• 利用状況に関する情報（アクセス日時、IPアドレス、ブラウザ情報）</Text>
            <Text>• Cookieおよび類似技術による情報</Text>
          </Stack>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            2. 情報の利用目的
          </Title>
          <Text mb="sm">収集した情報は以下の目的で利用します：</Text>
          <Stack gap="xs" pl="md">
            <Text>• 本サービスの提供・運営のため</Text>
            <Text>• ユーザー認証およびアカウント管理のため</Text>
            <Text>• ゲームのマッチング機能の提供のため</Text>
            <Text>• ユーザーサポートおよびお問い合わせへの対応のため</Text>
            <Text>• サービスの改善および新機能の開発のため</Text>
            <Text>• 利用規約違反や不正行為の防止のため</Text>
          </Stack>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            3. 情報の第三者提供
          </Title>
          <Text mb="sm">
            本サービスは、以下の場合を除き、ユーザーの個人情報を第三者に提供することはありません：
          </Text>
          <Stack gap="xs" pl="md">
            <Text>• ユーザーの同意がある場合</Text>
            <Text>• 法令に基づく場合</Text>
            <Text>• 人の生命、身体または財産の保護のために必要がある場合</Text>
            <Text>• 国の機関等への協力のために必要がある場合</Text>
          </Stack>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            4. Cookieおよび類似技術
          </Title>
          <Text>
            本サービスでは、ユーザーの利便性向上およびサービス改善のため、Cookieおよび類似技術を使用します。Cookieの使用を望まない場合は、ブラウザの設定により無効化することができますが、一部機能が制限される可能性があります。
          </Text>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            5. セキュリティ
          </Title>
          <Text>
            本サービスは、個人情報の漏洩、滅失または毀損の防止等、個人情報の安全管理のために必要かつ適切な措置を講じます。ただし、インターネット通信の特性上、セキュリティを完全に保証することはできません。
          </Text>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            6. 外部サービスの利用
          </Title>
          <Text>
            本サービスはGoogle認証サービスを利用しています。Google認証サービスの利用にあたっては、Googleのプライバシーポリシーが適用されます。詳細は
            <Text component="a" href="https://policies.google.com/privacy" target="_blank" rel="noopener noreferrer" c="blue" style={{ textDecoration: "underline" }}>
              Googleプライバシーポリシー
            </Text>
            をご確認ください。
          </Text>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            7. 個人情報の開示・訂正・削除
          </Title>
          <Text>
            ユーザーは、自己の個人情報について、開示、訂正、削除等を請求することができます。請求を希望される場合は、本サービスのお問い合わせ窓口までご連絡ください。
          </Text>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            8. プライバシーポリシーの変更
          </Title>
          <Text>
            本プライバシーポリシーは、法令の変更や本サービスの機能変更等に伴い、予告なく変更されることがあります。変更後のプライバシーポリシーは、本ページに掲載した時点で効力を生じるものとします。
          </Text>
        </div>

        <div>
          <Title order={2} size="h3" mb="md">
            9. お問い合わせ
          </Title>
          <Text>
            本プライバシーポリシーに関するお問い合わせは、本サービス内のお問い合わせフォームまたはサポート窓口までご連絡ください。
          </Text>
        </div>
      </Stack>
    </Container>
  );
}
