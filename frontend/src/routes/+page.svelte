<script lang="ts">
    import { onMount, getContext, type Component } from "svelte";
    import Clock from "$lib/components/Clock.svelte";

    const triggerReady = getContext<() => void>("app_ready");

    let AcademicProgress = $state<Component | null>(null);
    let SensorsWidget = $state<Component | null>(null);
    let InfoTicker = $state<Component | null>(null);

    onMount(() => {
        (async () => {
            const [sensorsModule, academicModule, tickerModule] =
                await Promise.all([
                    import("$lib/components/widgets/SensorsWidget.svelte"),
                    import("$lib/components/widgets/AcademicProgress.svelte"),
                    import("$lib/components/widgets/InfoTicker.svelte"),
                ]);
            SensorsWidget = sensorsModule.default;
            AcademicProgress = academicModule.default;
            InfoTicker = tickerModule.default;

            requestAnimationFrame(() => {
                requestAnimationFrame(() => {
                    triggerReady?.();
                });
            });
        })();
    });
</script>

<svelte:head>
    <title>Kiosk Display</title>
</svelte:head>

<main
    class="relative font-mono flex min-h-screen w-full items-center justify-center overflow-hidden select-none"
>
    {#if SensorsWidget}
        <SensorsWidget />
    {/if}

    {#if InfoTicker}
        <InfoTicker />
    {/if}

    <div class="flex flex-col items-center justify-center">
        <Clock />
        {#if AcademicProgress}
            <AcademicProgress />
        {/if}
    </div>
</main>
