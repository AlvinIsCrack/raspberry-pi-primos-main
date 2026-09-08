<script lang="ts">
    /**
     * Formatter configured strictly for 24-hour display in Santiago timezone.
     */
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

<time
    class="font-mono text-8xl font-bold tracking-tight text-white tabular-nums select-none"
>
    {currentTime}
</time>
