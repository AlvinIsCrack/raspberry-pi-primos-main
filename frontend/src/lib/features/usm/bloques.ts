import {
    getMinutesOfDay,
    isWithinMinutesRange,
    getMinutesIntervalProgress
} from '$lib/utils/time';

export const BLOCK_DURATION_MINUTES = 35;
export const BLOCK_INTERMISSION_DURATION_MINUTES = 15;
export const MODULE_TOTAL_DURATION_MINUTES = (BLOCK_DURATION_MINUTES * 2) + BLOCK_INTERMISSION_DURATION_MINUTES; // 85 min

export const LUNCH_START_MINUTES = 13 * 60 + 40; // 13:40 -> 820 min
export const LUNCH_END_MINUTES = 14 * 60 + 40;   // 14:40 -> 880 min

export const DAY_START_HOUR = 8;
export const DAY_START_MINUTE = 15;
export const TOTAL_BLOCKS = 10;

export enum PeriodKind {
    OffHours = 'OFF_HOURS',
    Lunch = 'LUNCH',
    Lecture = 'LECTURE',
    Intermission = 'INTERMISSION'
}

export interface AcademicBlock {
    firstIndex: number;
    secondIndex: number;
    startHour: number;
    startMin: number;
    endHour: number;
    endMin: number;
}

export interface CurrentScheduleState {
    kind: PeriodKind;
    block?: AcademicBlock;
    activeSubBlock?: 1 | 2;
    progress: number;
    minutesRange: [start: number, end: number] | null;
}

export function buildStandardBlocks(): AcademicBlock[] {
    const blocks: AcademicBlock[] = [];
    let cursor = DAY_START_HOUR * 60 + DAY_START_MINUTE;

    for (let i = 0; i < TOTAL_BLOCKS; i++) {
        const firstIdx = i * 2 + 1;
        const secondIdx = firstIdx + 1;

        if (cursor >= LUNCH_START_MINUTES && cursor < LUNCH_END_MINUTES) {
            cursor = LUNCH_END_MINUTES;
        }

        const endCursor = cursor + BLOCK_DURATION_MINUTES * 2;

        blocks.push({
            firstIndex: firstIdx,
            secondIndex: secondIdx,
            startHour: Math.floor(cursor / 60),
            startMin: cursor % 60,
            endHour: Math.floor(endCursor / 60),
            endMin: endCursor % 60
        });

        // Si el bloque termina justo donde empieza el almuerzo, no hay receso de 15 min
        if (endCursor === LUNCH_START_MINUTES) {
            cursor = LUNCH_END_MINUTES;
        } else {
            cursor = endCursor + BLOCK_INTERMISSION_DURATION_MINUTES;
        }
    }

    return blocks;
}

export const STANDARD_BLOCKS = buildStandardBlocks();

export function getCurrentAcademicSchedule(date: Date = new Date()): CurrentScheduleState {
    const current = getMinutesOfDay(date);

    // 1. Periodo de almuerzo
    if (isWithinMinutesRange(current, LUNCH_START_MINUTES, LUNCH_END_MINUTES, { inclusiveEnd: false })) {
        const progress = getMinutesIntervalProgress(current, LUNCH_START_MINUTES, LUNCH_END_MINUTES);
        return {
            kind: PeriodKind.Lunch,
            progress,
            minutesRange: [LUNCH_START_MINUTES, LUNCH_END_MINUTES]
        };
    }

    // 2. Módulos lectivos e intermisiones
    for (let i = 0; i < STANDARD_BLOCKS.length; i++) {
        const block = STANDARD_BLOCKS[i];
        const moduleStart = block.startHour * 60 + block.startMin;
        const lectureEnd = block.endHour * 60 + block.endMin;

        // Si el bloque termina a la hora de almuerzo, no tiene intermisión posterior
        const hasIntermission = lectureEnd !== LUNCH_START_MINUTES;
        const moduleEnd = hasIntermission
            ? lectureEnd + BLOCK_INTERMISSION_DURATION_MINUTES
            : lectureEnd;

        // ¿Está dentro del módulo (clases o su posterior descanso)?
        if (isWithinMinutesRange(current, moduleStart, moduleEnd, { inclusiveStart: true, inclusiveEnd: false })) {
            const progress = getMinutesIntervalProgress(current, moduleStart, moduleEnd);
            const minutesRange: [number, number] = [moduleStart, moduleEnd];

            // Sub-bloque 1 o 2 de clases
            if (current < lectureEnd) {
                const midpoint = moduleStart + BLOCK_DURATION_MINUTES;
                const activeSubBlock: 1 | 2 = current >= midpoint ? 2 : 1;

                return {
                    kind: PeriodKind.Lecture,
                    block,
                    activeSubBlock,
                    progress,
                    minutesRange: [moduleStart, lectureEnd]
                };
            }

            // Descanso / intermisión del módulo
            return {
                kind: PeriodKind.Intermission,
                block,
                progress,
                minutesRange
            };
        }
    }

    // 3. Fuera de horario
    return {
        kind: PeriodKind.OffHours,
        progress: 0,
        minutesRange: null
    };
}