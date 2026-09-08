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

    // Formateo derivado del título principal del módulo o periodo
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

    // Formateo derivado del rango horario del módulo completo (o almuerzo)
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
    <div class="flex flex-col items-center gap-1 mt-2 w-72 select-none">
        <!-- Barra de progreso -->
        <div
            class="relative w-full h-6 rounded bg-neutral-900 border border-neutral-800 overflow-hidden"
        >
            <div
                class="h-full transition-all duration-500 ease-out bg-neutral-200"
                style="width: {progressPercent}%"
            ></div>

            <div
                class="absolute left-1/2 top-1/2 -translate-1/2 font-semibold text-base text-white mix-blend-exclusion"
            >
                <span>{progressPercent}%</span>
            </div>
        </div>

        <!-- Información del módulo / bloque -->
        <div
            class="flex w-full items-center justify-between text-base text-neutral-400 font-mono"
        >
            <span class="tracking-tight text-neutral-400">
                <span>{label}</span>
                {#if currentSchedule.block}
                    {@const { firstIndex, secondIndex } = currentSchedule.block}
                    {@const active = currentSchedule.activeSubBlock}
                    {" "}
                    <span class="inline-flex items-center tracking-wide"
                        ><span
                            class:font-bold={active === 1}
                            class:text-white={active === 1}>{firstIndex}</span
                        >-<span
                            class:font-bold={active === 2}
                            class:text-white={active === 2}>{secondIndex}</span
                        ></span
                    >
                {/if}
            </span>
            <span class="text-neutral-500 text-base">
                {timeRange}
            </span>
        </div>
    </div>
{/if}
