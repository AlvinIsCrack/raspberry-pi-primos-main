import { TypedEventSource } from '$lib/api/sse-client';
import type { RoomCollection, RoomSnapshot } from '$lib/core/domain/room';

export interface RoomStreamEvents {
    room_updated: RoomSnapshot;
}

interface RawSensorSnapshot {
    room_id: string;
    door: string;
    battery_level: number;
    last_seen_at: string;
    connectivity?: string;
    is_stale?: boolean;
}

/**
 * Manages room snapshot fetching and synchronizes updates through real-time streams.
 */
export class RoomService {
    private readonly sseClient: TypedEventSource<RoomStreamEvents>;
    private activeWatchers = 0;

    constructor(
        private readonly apiBaseUrl: string = '/api/rooms',
        eventsBaseUrl: string = '/api/events'
    ) {
        this.sseClient = new TypedEventSource<RoomStreamEvents>(eventsBaseUrl);
    }

    public async fetchInitialSnapshots(): Promise<RoomCollection> {
        const response = await fetch(this.apiBaseUrl, {
            method: 'GET',
            headers: { Accept: 'application/json' }
        });

        if (!response.ok) {
            throw new Error(`Failed to load rooms: ${response.status} ${response.statusText}`);
        }

        const rawList = (await response.json()) as RawSensorSnapshot[];
        const collection: RoomCollection = {};

        for (const item of rawList) {
            const doorState = item.door === 'N/A' ? 'UNKNOWN' : (item.door as RoomSnapshot['door']);

            collection[item.room_id] = {
                id: item.room_id,
                door: doorState,
                batteryLevel: item.battery_level,
                lastSeen: item.last_seen_at
            };
        }

        return collection;
    }

    private mapToSnapshot(raw: RawSensorSnapshot): RoomSnapshot {
        const doorState = raw.door === 'N/A' || !raw.door ? 'UNKNOWN' : (raw.door as RoomSnapshot['door']);
        return {
            id: raw.room_id,
            door: doorState,
            batteryLevel: raw.battery_level,
            lastSeen: raw.last_seen_at
        };
    }

    public watchRooms(onUpdate: (room: RoomSnapshot) => void): () => void {
        if (this.activeWatchers === 0) {
            this.sseClient.connect();
        }
        this.activeWatchers++;

        const unsubscribe = this.sseClient.subscribe('room_updated', (data: unknown) => {
            onUpdate(this.mapToSnapshot(data as RawSensorSnapshot));
        });

        return () => {
            unsubscribe();
            this.activeWatchers--;
            if (this.activeWatchers === 0) {
                this.sseClient.disconnect();
            }
        };
    }
}
