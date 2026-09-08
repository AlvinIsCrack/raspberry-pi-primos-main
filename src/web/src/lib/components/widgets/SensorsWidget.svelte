<script lang="ts">
    import type { RoomCollection } from "$lib/core/domain/room";

    interface Props {
        rooms: RoomCollection;
    }

    let { rooms }: Props = $props();

    const roomList = $derived(Object.values(rooms));
</script>

<aside
    class="absolute top-6 right-6 z-20 flex flex-col gap-2 rounded-xl border border-neutral-800 bg-neutral-950/80 p-3.5 backdrop-blur-md min-w-[200px] shadow-2xl"
>
    <header
        class="flex items-center justify-between border-b border-neutral-800 pb-2"
    >
        <span
            class="text-[11px] font-semibold tracking-wider text-neutral-400 uppercase"
        >
            Rooms
        </span>
        <span class="font-mono text-[10px] text-neutral-500">
            {roomList.length} active
        </span>
    </header>

    {#if roomList.length === 0}
        <p class="py-2 text-center text-xs text-neutral-600">
            No sensors connected
        </p>
    {:else}
        <ul class="flex flex-col gap-1.5">
            {#each roomList as room (room.id)}
                <li
                    class="flex items-center justify-between gap-3 rounded-lg bg-neutral-900/50 px-2.5 py-1.5 border border-neutral-850"
                >
                    <div class="flex items-center gap-2">
                        <span
                            class="h-2 w-2 rounded-full ring-2 ring-neutral-950"
                            class:bg-emerald-500={room.door === "CLOSED"}
                            class:bg-red-500={room.door === "OPEN"}
                            class:bg-amber-400={room.door === "UNKNOWN"}
                            class:animate-pulse={room.door === "UNKNOWN"}
                        ></span>
                        <span
                            class="font-mono text-xs font-bold text-neutral-200"
                        >
                            {room.id}
                        </span>
                    </div>
                    <div class="flex items-center gap-2">
                        <span
                            class="text-[10px] font-medium tracking-wide uppercase"
                            class:text-neutral-400={room.door !== "UNKNOWN"}
                            class:text-amber-400={room.door === "UNKNOWN"}
                            class:animate-pulse={room.door === "UNKNOWN"}
                        >
                            {room.door === "UNKNOWN" ? "N/A" : room.door}
                        </span>
                        <span
                            class="font-mono text-[11px] text-neutral-500 tabular-nums"
                        >
                            {room.door === "UNKNOWN"
                                ? "--"
                                : `${room.batteryLevel}%`}
                        </span>
                    </div>
                </li>
            {/each}
        </ul>
    {/if}
</aside>
