import {
    getMinutesOfDay,
    isWithinMinutesRange,
    minutesToTimeString
} from '$lib/utils/time';

export const BLOCK_DURATION_MINUTES = 35;
export const BLOCK_INTERMISSION_DURATION_MINUTES = 15;
export const LUNCH_START_MINUTES = 13 * 60 + 40; // 13:40 -> 820 min
export const LUNCH_END_MINUTES = 14 * 60 + 30;   // 14:30 -> 870 min

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
}

/**
 * Genera de forma determinista la grilla estándar de bloques dobles de la USM.
 */
export function buildStandardBlocks(): AcademicBlock[] {
    const blocks: AcademicBlock[] = [];
    let cursor = DAY_START_HOUR * 60 + DAY_START_MINUTE;

    for (let i = 0; i < TOTAL_BLOCKS; i++) {
        const firstIdx = i * 2 + 1;
        const secondIdx = firstIdx + 1;

        // Salta el intervalo de almuerzo institucional si coincide
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

        cursor = endCursor + BLOCK_INTERMISSION_DURATION_MINUTES;
    }

    return blocks;
}

export const STANDARD_BLOCKS = buildStandardBlocks();

/**
 * Evalúa el momento temporal contra los bloques calculados, descansos y almuerzo.
 */
export function getCurrentAcademicSchedule(date: Date = new Date()): CurrentScheduleState {
    // Obtiene los minutos del día respetando la zona horaria (America/Santiago por defecto)
    const current = getMinutesOfDay(date);

    // Horario de almuerzo
    if (isWithinMinutesRange(current, LUNCH_START_MINUTES, LUNCH_END_MINUTES, { inclusiveEnd: false })) {
        return { kind: PeriodKind.Lunch };
    }

    for (let i = 0; i < STANDARD_BLOCKS.length; i++) {
        const block = STANDARD_BLOCKS[i];
        const start = block.startHour * 60 + block.startMin;
        const end = block.endHour * 60 + block.endMin;

        // En clase / periodo lectivo (inclusive ambos extremos según la lógica original)
        if (isWithinMinutesRange(current, start, end, { inclusiveStart: true, inclusiveEnd: false })) {
            const midpoint = start + BLOCK_DURATION_MINUTES;
            return {
                kind: PeriodKind.Lecture,
                block,
                activeSubBlock: current >= midpoint ? 2 : 1
            };
        }

        // En ventana de descanso (intermisión entre bloques)
        if (i < STANDARD_BLOCKS.length - 1) {
            const nextStart = STANDARD_BLOCKS[i + 1].startHour * 60 + STANDARD_BLOCKS[i + 1].startMin;
            if (isWithinMinutesRange(current, end, nextStart, { inclusiveStart: true, inclusiveEnd: false })) {
                return {
                    kind: PeriodKind.Intermission,
                    block
                };
            }
        }
    }

    return { kind: PeriodKind.OffHours };
}