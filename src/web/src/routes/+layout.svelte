<script lang="ts">
	import "$lib/css/layout.css";
	import BootSequence from "$lib/components/BootSequence.svelte";
	import { setContext } from "svelte";

	let { children } = $props();
	let isBooting = $state(true);
	let isReady = $state(false);

	setContext("app_ready", () => {
		isReady = true;
	});
</script>

<svelte:head>
	<link rel="icon" href="/favicon.svg" />
</svelte:head>

{#if isBooting}
	<BootSequence ready={isReady} oncomplete={() => (isBooting = false)} />
{/if}

{@render children?.()}
