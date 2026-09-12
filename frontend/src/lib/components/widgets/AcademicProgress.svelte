<script lang="ts">
    import { AcademicScheduleManager } from "$lib/features/usm/schedule.svelte";
    import {
        PeriodKind,
        LUNCH_START_MINUTES,
        LUNCH_END_MINUTES,
        BLOCK_INTERMISSION_DURATION_MINUTES,
    } from "$lib/features/usm/bloques";
    import { minutesToTimeString } from "$lib/utils/time";
    import { onDestroy } from "svelte";

    const schedule = new AcademicScheduleManager();
    const currentSchedule = $derived(schedule.current);
    const isVisible = $derived(currentSchedule.kind !== PeriodKind.OffHours);
    const progressPercent = $derived(
        Math.round((currentSchedule.progress ?? 0) * 100),
    );

    // Estado para controlar la animación de transición de módulo
    let isTransitioning = $state(false);
    let transitionLabel = $state("");
    let lastModuleKey = $state<string | null>(null);

    // Detección real basada en cambio de identificador de periodo/bloque
    $effect(() => {
        const kind = currentSchedule.kind;
        const firstIdx = currentSchedule.block?.firstIndex ?? 0;
        const secondIdx = currentSchedule.block?.secondIndex ?? 0;

        // Clave única que representa el bloque o periodo actual
        const currentKey = `${kind}-${firstIdx}-${secondIdx}`;

        // Si ya teníamos un bloque registrado y este es diferente, ocurrió un cambio
        if (lastModuleKey !== null && lastModuleKey !== currentKey) {
            transitionLabel = getModuleTitle(currentSchedule);
            isTransitioning = true;

            const timer = setTimeout(() => {
                isTransitioning = false;
            }, 3000); // Duración de la animación en grande

            return () => clearTimeout(timer);
        }

        // Actualizamos la referencia del último módulo conocido
        lastModuleKey = currentKey;
    });

    function getModuleTitle(sch: typeof currentSchedule): string {
        switch (sch.kind) {
            case PeriodKind.Lecture:
                return sch.block
                    ? `Bloque ${sch.block.firstIndex}-${sch.block.secondIndex}`
                    : "Bloque";
            case PeriodKind.Intermission:
                return "Receso";
            case PeriodKind.Lunch:
                return "Almuerzo";
            default:
                return "Fuera de horario";
        }
    }

    const label = $derived.by(() => {
        switch (currentSchedule.kind) {
            case PeriodKind.Lecture:
                return "Bloque";
            case PeriodKind.Intermission:
                return "Receso";
            case PeriodKind.Lunch:
                return "Almuerzo";
            default:
                return "";
        }
    });

    const timeRange = $derived.by(() => {
        const { minutesRange } = currentSchedule;
        if (!minutesRange || minutesRange.length < 2) return "";
        return `${minutesToTimeString(minutesRange[0])}-${minutesToTimeString(minutesRange[1])}`;
    });

    onDestroy(() => {
        schedule.stop();
    });
</script>

{#if isVisible}
    <div
        class="relative flex flex-col items-center gap-1 mt-2 w-sm select-none"
    >
        <div
            class="absolute inset-0 z-10 flex items-center justify-center bg-background transition-opacity duration-700 pointer-events-none"
            class:opacity-100={isTransitioning}
            class:opacity-0={!isTransitioning}
        >
            <span
                class="text-4xl -mt-4 font-bold tracking-wider uppercase text-foreground animate-blink"
            >
                {transitionLabel}
            </span>
        </div>

        <div
            class="w-full flex flex-col gap-1 transition-opacity duration-700"
            class:opacity-0={isTransitioning}
            class:opacity-100={!isTransitioning}
        >
            <!-- Barra de progreso -->
            <div
                class="relative w-full h-8 rounded bg-background border overflow-hidden"
            >
                <div
                    class="h-full bg-border"
                    style="width: {progressPercent}%"
                ></div>
                <div
                    class="absolute left-1/2 top-1/2 -translate-1/2 font-semibold text-base text-foreground bg-background rounded px-2"
                >
                    <span>{progressPercent}%</span>
                </div>
            </div>

            <!-- Información del módulo / bloque -->
            <div
                class="flex w-full items-center justify-between text-base text-neutral-400 font-mono"
            >
                <span class="tracking-tight text-muted-content">
                    <span>{label}</span>
                    {#if currentSchedule.block}
                        {@const { firstIndex, secondIndex } =
                            currentSchedule.block}
                        {@const active = currentSchedule.activeSubBlock}
                        {" "}
                        <span class="inline-flex items-center tracking-wide">
                            <span
                                class:font-bold={active === 1}
                                class:text-foreground={active === 1}
                                >{firstIndex}</span
                            >-<span
                                class:font-bold={active === 2}
                                class:text-foreground={active === 2}
                                >{secondIndex}</span
                            >
                        </span>
                    {/if}
                </span>
                <span class="text-muted-content text-base">
                    {timeRange}
                </span>
            </div>
        </div>
    </div>
{/if}
