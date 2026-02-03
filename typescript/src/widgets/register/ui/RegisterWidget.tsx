import { Button, Container, Paper, Text, Title, Stack } from "@mantine/core";
import Link from "next/link";
import { IconBrandGoogle } from "@tabler/icons-react";

export function RegisterWidget() {
  return (
    <Container size={420} my={40}>
      <Title ta="center" fw={900}>
        新規登録
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5}>
        既にアカウントをお持ちの場合は{" "}
        <Link href="/login" style={{ color: "var(--mantine-color-blue-filled)" }}>
          ログイン
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
            Googleで登録
          </Button>
        </Stack>
      </Paper>
    </Container>
  );
}
