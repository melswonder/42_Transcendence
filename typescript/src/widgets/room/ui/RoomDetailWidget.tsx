interface RoomDetailWidgetProps {
  roomId: string;
}

export function RoomDetailWidget({ roomId }: RoomDetailWidgetProps) {
  return (
    <div>
      <h1>ルーム ID: {roomId}</h1>
    </div>
  );
}
