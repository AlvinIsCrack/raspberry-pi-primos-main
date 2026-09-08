<script lang="ts">
    import Clock from "$lib/components/Clock.svelte";
    import AcademicProgress from "$lib/components/widgets/AcademicProgress.svelte";
    import SensorsWidget from "$lib/components/widgets/SensorsWidget.svelte";
</script>

<svelte:head>
    <title>Kiosk Display</title>
</svelte:head>

<main
    class="relative font-mono flex min-h-screen w-full items-center justify-center overflow-hidden select-none"
>
    <SensorsWidget />
    <div class="flex flex-col items-center justify-center">
        <Clock />
        <AcademicProgress />
    </div>
</main>
