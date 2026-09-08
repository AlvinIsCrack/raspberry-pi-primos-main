export type TimeZone = string;

export interface TimeParts {
    hour: number;
    minute: number;
    second: number;
}

export interface FormatTimeOptions {
    timeZone?: TimeZone;
    locale?: string;
    includeSeconds?: boolean;
    hour12?: boolean;
}

export interface MinutesRangeOptions {
    inclusiveStart?: boolean;
    inclusiveEnd?: boolean;
    /** Maneja rangos que cruzan la medianoche (ej. 23:00 a 02:00) */
    wrapAroundMidnight?: boolean;
}

// Configuración por defecto personalizable
export const DEFAULT_TIMEZONE: TimeZone = 'America/Santiago';
export const DEFAULT_LOCALE = 'es-CL';
export const MINUTES_IN_DAY = 1440;

// Caché de formatters para alto rendimiento
const formatterCache = new Map<string, Intl.DateTimeFormat>();

function getOrCreateFormatter(
    timeZone: TimeZone = DEFAULT_TIMEZONE,
    locale: string = DEFAULT_LOCALE,
    options: Intl.DateTimeFormatOptions = {}
): Intl.DateTimeFormat {
    const key = `${locale}_${timeZone}_${JSON.stringify(options)}`;
    let formatter = formatterCache.get(key);
    if (!formatter) {
        formatter = new Intl.DateTimeFormat(locale, {
            timeZone,
            ...options
        });
        formatterCache.set(key, formatter);
    }
    return formatter;
}

/**
 * Extrae las partes horarias numéricas respetando cualquier zona horaria IANA.
 */
export function getTimeParts(
    date: Date = new Date(),
    timeZone: TimeZone = DEFAULT_TIMEZONE,
    locale: string = DEFAULT_LOCALE
): TimeParts {
    const formatter = getOrCreateFormatter(timeZone, locale, {
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric',
        hourCycle: 'h23' // Garantiza rango 0-23
    });

    const parts = formatter.formatToParts(date);
    let hour = 0;
    let minute = 0;
    let second = 0;

    for (const part of parts) {
        if (part.type === 'hour') hour = parseInt(part.value, 10);
        else if (part.type === 'minute') minute = parseInt(part.value, 10);
        else if (part.type === 'second') second = parseInt(part.value, 10);
    }

    return { hour: hour % 24, minute, second };
}

/**
 * Retorna los minutos transcurridos desde la medianoche (0 - 1439).
 */
export function getMinutesOfDay(
    date: Date = new Date(),
    timeZone: TimeZone = DEFAULT_TIMEZONE
): number {
    const { hour, minute } = getTimeParts(date, timeZone);
    return hour * 60 + minute;
}

/**
 * Alias retrocompatible con la firma anterior.
 */
export const getMinutes = getMinutesOfDay;

/**
 * Convierte una cadena horaria "HH:mm" o "HH:mm:ss" a minutos del día.
 * Retorna null si el formato es inválido.
 */
export function parseTimeStringToMinutes(timeStr: string): number | null {
    const match = /^(\d{1,2}):(\d{2})(?::(\d{2}))?$/.exec(timeStr.trim());
    if (!match) return null;

    const hour = parseInt(match[1], 10);
    const minute = parseInt(match[2], 10);

    if (hour < 0 || hour > 23 || minute < 0 || minute > 59) return null;

    return hour * 60 + minute;
}

/**
 * Convierte minutos del día a una cadena "HH:mm" legible.
 */
export function minutesToTimeString(minutesOfDay: number, pad = true): string {
    const normalized = ((minutesOfDay % MINUTES_IN_DAY) + MINUTES_IN_DAY) % MINUTES_IN_DAY;
    const hours = Math.floor(normalized / 60);
    const minutes = normalized % 60;

    const hStr = pad ? hours.toString().padStart(2, '0') : hours.toString();
    const mStr = minutes.toString().padStart(2, '0');

    return `${hStr}:${mStr}`;
}

/**
 * Verifica si un minuto del día está dentro de un rango determinado.
 * Soporta rangos que traspasan la medianoche (ej: 22:00 a 06:00).
 */
export function isWithinMinutesRange(
    currentMinutes: number,
    startMinutes: number,
    endMinutes: number,
    options: MinutesRangeOptions = {}
): boolean {
    const { inclusiveStart = true, inclusiveEnd = true, wrapAroundMidnight = true } = options;

    if (wrapAroundMidnight && startMinutes > endMinutes) {
        // Intervalo nocturno (ej: de 23:00 [1380] a 04:00 [240])
        const inFirstPart = inclusiveStart ? currentMinutes >= startMinutes : currentMinutes > startMinutes;
        const inSecondPart = inclusiveEnd ? currentMinutes <= endMinutes : currentMinutes < endMinutes;
        return inFirstPart || inSecondPart;
    }

    const afterStart = inclusiveStart ? currentMinutes >= startMinutes : currentMinutes > startMinutes;
    const beforeEnd = inclusiveEnd ? currentMinutes <= endMinutes : currentMinutes < endMinutes;

    return afterStart && beforeEnd;
}

/**
 * Calcula el progreso relativo (0.0 a 1.0) dentro de un intervalo en minutos.
 */
export function getMinutesIntervalProgress(
    currentMinutes: number,
    startMinutes: number,
    endMinutes: number
): number {
    if (endMinutes <= startMinutes) return 0;
    if (currentMinutes <= startMinutes) return 0;
    if (currentMinutes >= endMinutes) return 1;

    return (currentMinutes - startMinutes) / (endMinutes - startMinutes);
}

/**
 * Retorna los milisegundos restantes hasta el cambio exacto del próximo minuto.
 * Ideal para sincronizar timers y loops de reloj exactos sin desvíos.
 */
export function msUntilNextMinute(now: Date = new Date()): number {
    return (60 - now.getSeconds()) * 1000 - now.getMilliseconds();
}

/**
 * Formatea una fecha a string de visualización según locale y zona horaria.
 */
export function formatTimeDisplay(
    date: Date = new Date(),
    options: FormatTimeOptions = {}
): string {
    const {
        timeZone = DEFAULT_TIMEZONE,
        locale = DEFAULT_LOCALE,
        includeSeconds = false,
        hour12 = false
    } = options;

    const formatter = getOrCreateFormatter(timeZone, locale, {
        hour: '2-digit',
        minute: '2-digit',
        second: includeSeconds ? '2-digit' : undefined,
        hour12
    });

    return formatter.format(date);
}