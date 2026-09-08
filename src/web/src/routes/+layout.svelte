<script lang="ts">
	import "$lib/css/layout.css";
	import BootSequence from "$lib/components/BootSequence.svelte";

	let { children } = $props();

	let isBooting = $state(true);
</script>

<svelte:head>
	<link rel="icon" href="/favicon.svg" />
</svelte:head>

{#if isBooting}
	<BootSequence oncomplete={() => (isBooting = false)} />
{:else}
	{@render children()}
{/if}
