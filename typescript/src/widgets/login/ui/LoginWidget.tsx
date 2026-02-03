import { Button, Container, Paper, Text, Title, Stack } from "@mantine/core";
import Link from "next/link";
import { IconBrandGoogle } from "@tabler/icons-react";

export function LoginWidget() {
  return (
    <Container size={420} my={40}>
      <Title ta="center" fw={900}>
        ログイン
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5}>
        アカウントをお持ちでない場合は{" "}
        <Link href="/register" style={{ color: "var(--mantine-color-blue-filled)" }}>
          新規登録
        </Link>
      </Text>

      <Paper withBorder shadow="md" p={30} mt={30} radius="md">
        <Stack>
          <Button
            leftSection={<IconBrandGoogle size={20} />}
            variant="default"
            fullWidth
            size="md"
          >
            Googleでログイン
          </Button>
        </Stack>
      </Paper>
    </Container>
  );
}
