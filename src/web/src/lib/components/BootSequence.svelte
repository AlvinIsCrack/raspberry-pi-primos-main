<script lang="ts">
    import { onMount } from "svelte";

    interface Props {
        ready?: boolean;
        oncomplete?: () => void;
    }

    let { ready = true, oncomplete }: Props = $props();

    let isVisible = $state(false);
    let isFadingOut = $state(false);
    let minTimePassed = $state(false);

    onMount(() => {
        // Fade in inicial del logo
        const showTimer = setTimeout(() => {
            isVisible = true;
        }, 50);

        // Tiempo mínimo garantizado para ver el logo (ej: 1.5s)
        const minDisplayTimer = setTimeout(() => {
            minTimePassed = true;
        }, 1500);

        return () => {
            clearTimeout(showTimer);
            clearTimeout(minDisplayTimer);
        };
    });

    // Solo inicia el fade-out cuando ya cargó todo Y pasó el tiempo mínimo
    $effect(() => {
        if (ready && minTimePassed && !isFadingOut) {
            isFadingOut = true;

            // 650ms coincide con la duración de la transición CSS
            const finishTimer = setTimeout(() => {
                oncomplete?.();
            }, 650);

            return () => clearTimeout(finishTimer);
        }
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
        class:opacity-0={!isVisible}
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
