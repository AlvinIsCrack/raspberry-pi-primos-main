<script lang="ts">
    import { RoomService } from "$lib/core/services/room.service";
    import type { RoomCollection } from "$lib/core/domain/room";

    const roomService = new RoomService();
    let rooms = $state<RoomCollection>({});
    let activityTicks = $state<Record<string, number>>({});

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
            activityTicks = { ...activityTicks, [updatedRoom.id]: Date.now() };
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
    class="absolute top-6 right-6 z-20 flex flex-col gap-2 rounded border-2 p-2 min-w-60"
>
    <div class="relative size-full">
        <div
            class="absolute -top-0.5 -translate-y-full text-border bg-background px-2"
        >
            Puertas
        </div>

        {#if roomList.length === 0}
            <p class="py-2 text-center text-sm text-border/80">
                No hay sensores registrados
            </p>
        {:else}
            {@const icons: Record<string, string> = {
                UNKNOWN: "*",
            }}
            <ul class="flex flex-col gap-1.5">
                {#each roomList as room (room.id)}
                    <li
                        class="relative overflow-hidden flex items-center justify-between gap-3 rounded px-4 py-2 border border-border/60"
                    >
                        <div class="flex items-center text-2xl gap-2">
                            {#if room.door in icons}
                                <span
                                    class="-mx-1 font-bold text-xl tracking-tighter"
                                    class:text-primary={room.door === "ABR"}
                                    class:text-secondary={room.door === "CER"}
                                    class:text-warning={room.door === "UNKNOWN"}
                                    class:animate-blink={room.door ===
                                        "UNKNOWN"}
                                >
                                    {icons[room.door]}
                                </span>
                            {/if}

                            <span class="font-mono font-bold text-neutral-200">
                                {room.id}
                            </span>

                            {#if activityTicks[room.id]}
                                {#key activityTicks[room.id]}
                                    <span
                                        class="activity-dot inline-block size-1.5 rounded-full bg-foreground"
                                    ></span>
                                {/key}
                            {/if}
                        </div>
                        <div class="flex items-center gap-2">
                            <span
                                class="text-xl font-medium tracking-wide uppercase"
                                class:text-primary={room.door === "ABR"}
                                class:text-secondary={room.door === "CER"}
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

<style>
    @keyframes fadeOut {
        0% {
            opacity: 1;
        }
        100% {
            opacity: 0;
        }
    }

    .activity-dot {
        animation: fadeOut 5s ease-out forwards;
    }
</style>
