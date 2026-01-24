import { BattleWidget } from '@/widgets/battle';

interface BattlePageProps {
  params: Promise<{ id: string }>;
}

export default async function BattlePage({ params }: BattlePageProps) {
  const { id } = await params;
  return <BattleWidget battleId={id} />;
}
