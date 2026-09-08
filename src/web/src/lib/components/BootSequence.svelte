<script lang="ts">
    import { onMount } from "svelte";

    interface Props {
        oncomplete?: () => void;
    }

    let { oncomplete }: Props = $props();

    // Estados de animación
    let isVisible = $state(false); // Controla el Fade In del logo
    let isFadingOut = $state(false); // Controla el Fade Out de toda la pantalla al negro

    onMount(() => {
        const showTimer = setTimeout(() => {
            isVisible = true;
        }, 80);

        const fadeOutTimer = setTimeout(() => {
            isFadingOut = true;
        }, 3800);

        const completeTimer = setTimeout(() => {
            oncomplete?.();
        }, 4400);

        return () => {
            clearTimeout(showTimer);
            clearTimeout(fadeOutTimer);
            clearTimeout(completeTimer);
        };
    });
</script>

<!-- Contenedor general en negro -->
<div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black select-none overflow-hidden transition-opacity duration-600 ease-in-out"
    class:opacity-0={isFadingOut}
    class:pointer-events-none={isFadingOut}
>
    <div
        class="flex flex-col items-center justify-center transition-all duration-700 ease-out transform"
        class:opacity-100={isVisible}
        class:scale-100={isVisible}
        class:opacity-0={!isVisible}
        class:scale-95={!isVisible}
    >
        <img
            class="logo-img h-50 w-auto object-contain flex items-center justify-center"
            alt=""
            src="/media/os-logo.png"
        />
    </div>
</div>

<style>
    .logo-img {
        image-rendering: -moz-crisp-edges;
        image-rendering: pixelated;
    }
</style>
