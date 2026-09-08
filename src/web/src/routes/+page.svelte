<script lang="ts">
    import Clock from "$lib/components/Clock.svelte";
    import SensorsWidget from "$lib/components/widgets/SensorsWidget.svelte";
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
            rooms[updatedRoom.id] = updatedRoom;
        });

        return () => {
            unsubscribe();
        };
    });
</script>

<svelte:head>
    <title>Kiosk Display</title>
</svelte:head>

<main
    class="relative flex min-h-screen w-full items-center justify-center bg-black overflow-hidden select-none"
>
    <SensorsWidget {rooms} />
    <Clock />
</main>
