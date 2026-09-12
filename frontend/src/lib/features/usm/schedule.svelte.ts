import { msUntilNextMinute } from '$lib/utils/time';
import { getCurrentAcademicSchedule, type CurrentScheduleState } from './bloques';

export class AcademicScheduleManager {
    current = $state<CurrentScheduleState>(getCurrentAcademicSchedule());
    private timerId?: ReturnType<typeof setTimeout>;

    constructor(autoStart = true) {
        if (autoStart && typeof window !== 'undefined') {
            this.start();
        }
    }

    public update() {
        this.current = getCurrentAcademicSchedule();
    }

    public start() {
        this.stop();
        this.update();

        const loop = () => {
            this.update();
            this.timerId = setTimeout(loop, msUntilNextMinute());
        };

        this.timerId = setTimeout(loop, msUntilNextMinute());
    }

    public stop() {
        if (this.timerId) {
            clearTimeout(this.timerId);
            this.timerId = undefined;
        }
    }
}