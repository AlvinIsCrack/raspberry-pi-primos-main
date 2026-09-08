<script lang="ts">
    import { RoomService } from "$lib/core/services/room.service";
    import type { RoomCollection } from "$lib/core/domain/room";

    const roomService = new RoomService();
    let rooms = $state<RoomCollection>({});

    $effect(() => {
        roomService
            .fetchInitialSnapshots()
            .then((data) => {
                rooms = data;
            })
            .catch(() => {
                // Snapshot retrieval failure handled by subsequent live stream updates
            });

        const unsubscribe = roomService.watchRooms((updatedRoom) => {
            rooms = { ...rooms, [updatedRoom.id]: updatedRoom };
        });

        return () => {
            unsubscribe();
        };
    });

    const roomList = $derived(
        Object.values(rooms).sort((a, b) => a.id.localeCompare(b.id)),
    );
</script>

<aside
    class="absolute top-6 right-6 z-20 flex flex-col gap-2 rounded border-2 border-neutral-600 p-2 min-w-60"
>
    <div class="relative size-full">
        <div
            class="absolute -top-0.5 -translate-y-full text-sm text-neutral-400 bg-black px-2"
        >
            Puertas
        </div>

        {#if roomList.length === 0}
            <p class="py-2 text-center text-xs text-neutral-600">
                No hay sensores registrados
            </p>
        {:else}
            <ul class="flex flex-col gap-1.5">
                {#each roomList as room (room.id)}
                    <li
                        class="relative overflow-hidden flex items-center justify-between gap-3 rounded bg-linear-to-r from-neutral-900/50 px-3 py-2 pl-8 border border-neutral-600/40"
                    >
                        <div class="flex items-center gap-2">
                            <div
                                class="h-full absolute left-0 w-5 -z-10"
                                class:bg-secondary={room.door === "CLOSED"}
                                class:bg-primary={room.door === "OPEN"}
                                class:bg-warning={room.door === "UNKNOWN"}
                                class:animate-blink={room.door === "UNKNOWN"}
                            ></div>
                            <span
                                class="font-mono text-base font-bold text-neutral-200"
                            >
                                {room.id}
                            </span>
                        </div>
                        <div class="flex items-center gap-2">
                            <span
                                class="text-sm font-medium tracking-wide uppercase"
                                class:text-neutral-400={room.door !== "UNKNOWN"}
                                class:text-warning={room.door === "UNKNOWN"}
                                class:animate-blink={room.door === "UNKNOWN"}
                            >
                                {room.door === "UNKNOWN" ? "N/A" : room.door}
                            </span>
                        </div>
                    </li>
                {/each}
            </ul>
        {/if}
    </div>
</aside>
