<script lang="ts">
    import { formatTimeDisplay, msUntilNextMinute } from "$lib/utils/time";

    let currentTime = $state(formatTimeDisplay());

    $effect(() => {
        let timer: ReturnType<typeof setTimeout>;

        const tick = () => {
            currentTime = formatTimeDisplay();
            // Calcula el retraso exacto hasta el cambio de minuto para evitar desvíos
            timer = setTimeout(tick, msUntilNextMinute());
        };

        tick();
        return () => clearTimeout(timer);
    });
</script>

<time
    class="font-mono text-[8.6rem] leading-none font-bold tracking-tight text-foreground -mb-2 tabular-nums select-none"
>
    {currentTime}
</time>
