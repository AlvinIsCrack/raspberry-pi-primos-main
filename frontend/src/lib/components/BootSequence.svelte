<script lang="ts">
    import { onMount } from "svelte";

    interface Props {
        ready?: boolean;
        oncomplete?: () => void;
    }

    interface SystemStatus {
        connected: boolean;
        ip?: string;
    }

    let { ready = true, oncomplete }: Props = $props();
    let isVisible = $state(false);
    let isFadingOut = $state(false);
    let minTimePassed = $state(false);
    let systemStatus = $state<SystemStatus | null>(null);

    onMount(() => {
        // Consultar el estado del sistema al arrancar
        fetch("/api/system/status")
            .then((res) => (res.ok ? res.json() : Promise.reject()))
            .then((data: { connected?: boolean; local_ips?: string[] }) => {
                systemStatus = {
                    connected: Boolean(data.connected),
                    ip: data.local_ips?.[0],
                };
            })
            .catch(() => {
                systemStatus = { connected: false };
            });

        // Fade-in inicial
        const showTimer = setTimeout(() => {
            isVisible = true;
        }, 50);

        // Tiempo mínimo en pantalla aumentado a 5 segundos
        const minDisplayTimer = setTimeout(() => {
            minTimePassed = true;
        }, 5000);

        return () => {
            clearTimeout(showTimer);
            clearTimeout(minDisplayTimer);
        };
    });

    // Iniciar fade-out al cumplirse los 5 segundos y estar la app lista
    $effect(() => {
        if (ready && minTimePassed && !isFadingOut) {
            isFadingOut = true;
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
    <!-- Logo central -->
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

    <!-- Indicador de red e IP en la esquina inferior izquierda -->
    {#if systemStatus}
        <div
            class="absolute bottom-6 left-6 flex items-center gap-2 font-mono text-sm tracking-wide transition-opacity duration-700 ease-out"
            class:opacity-100={isVisible}
            class:opacity-0={!isVisible}
        >
            <span
                class="size-2 rounded-full"
                class:bg-success={systemStatus.connected}
                class:bg-danger={!systemStatus.connected}
            ></span>
            {#if systemStatus.connected && systemStatus.ip}
                <span class="text-muted-content font-semibold"
                    >{systemStatus.ip}</span
                >
            {/if}
        </div>
    {/if}
</div>

<style>
    .logo-img {
        image-rendering: -moz-crisp-edges;
        image-rendering: pixelated;
    }
</style>
