export type DoorState = 'OPEN' | 'CLOSED' | 'UNKNOWN';

export interface RoomSnapshot {
    id: string;
    door: DoorState;
    batteryLevel: number;
    lastSeen: string;
}

export type RoomCollection = Record<string, RoomSnapshot>;