<script lang="ts">
    import Clock from "$lib/components/Clock.svelte";
    import SensorsWidget from "$lib/components/widgets/SensorsWidget.svelte";
</script>

<svelte:head>
    <title>Kiosk Display</title>
</svelte:head>

<main
    class="relative font-mono flex min-h-screen w-full items-center justify-center bg-black overflow-hidden select-none"
>
    <SensorsWidget />
    <Clock />
</main>
