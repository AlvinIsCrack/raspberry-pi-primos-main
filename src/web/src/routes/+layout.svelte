<script lang="ts">
	import "$lib/css/layout.css";

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href="/favicon.svg" />
</svelte:head>
{@render children()}
