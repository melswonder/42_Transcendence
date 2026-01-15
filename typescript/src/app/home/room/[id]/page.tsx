import { RoomDetailWidget } from '@/widgets/room';

export default function RoomDetailPage({ params }: { params: { id: string } }) {
  return <RoomDetailWidget roomId={params.id} />;
}
