import { TypedEventSource } from '$lib/api/sse-client';
import type { RoomCollection, RoomSnapshot } from '$lib/core/domain/room';

export interface RoomStreamEvents {
    room_updated: RoomSnapshot;
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

        return (await response.json()) as RoomCollection;
    }

    public watchRooms(onUpdate: (room: RoomSnapshot) => void): () => void {
        if (this.activeWatchers === 0) {
            this.sseClient.connect();
        }
        this.activeWatchers++;

        const unsubscribe = this.sseClient.subscribe('room_updated', onUpdate);

        return () => {
            unsubscribe();
            this.activeWatchers--;
            if (this.activeWatchers === 0) {
                this.sseClient.disconnect();
            }
        };
    }
}
