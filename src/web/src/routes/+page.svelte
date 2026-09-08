<script lang="ts">
    const timeFormatter = new Intl.DateTimeFormat("es-CL", {
        timeZone: "America/Santiago",
        hour: "2-digit",
        minute: "2-digit",
        hour12: false,
    });

    let currentTime = $state(timeFormatter.format(new Date()));

    $effect(() => {
        let timer: ReturnType<typeof setTimeout>;

        const tick = () => {
            const now = new Date();
            currentTime = timeFormatter.format(now);
            const delay = 1000 - now.getMilliseconds();
            timer = setTimeout(tick, delay);
        };

        tick();
        return () => clearTimeout(timer);
    });
</script>

<svelte:head>
    <title>Kiosk Display</title>
</svelte:head>

<main
    class="flex min-h-screen w-full flex-col items-center justify-center bg-black select-none"
>
    <time
        class="font-mono text-7xl font-bold tracking-tight text-white tabular-nums"
    >
        {currentTime}
    </time>
</main>
