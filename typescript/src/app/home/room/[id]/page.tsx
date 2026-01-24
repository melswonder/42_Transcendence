import { RoomDetailWidget } from '@/widgets/room';

interface RoomPageProps {
  params: Promise<{ id: string }>;
}

export default async function RoomPage({ params }: RoomPageProps) {
  const { id } = await params;
  return <RoomDetailWidget roomId={id} />;
}
