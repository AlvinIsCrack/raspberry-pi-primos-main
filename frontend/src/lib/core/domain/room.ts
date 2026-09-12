export type DoorState = 'ABR' | 'CER' | 'UNKNOWN';

export interface RoomSnapshot {
    id: string;
    door: DoorState;
    batteryLevel: number;
    lastSeen: string;
}

export type RoomCollection = Record<string, RoomSnapshot>;

export function formatDoorStateLabel(door: RoomSnapshot['door']): string {
    if (door === 'ABR') return 'O';
    if (door === 'CER') return 'X';
    return 'N/A';
}