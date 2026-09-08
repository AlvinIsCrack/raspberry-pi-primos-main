export type SSEEventHandler<T> = (data: T) => void;
export type SSEErrorHandler = (error: Event) => void;

/**
 * Type-safe abstraction over the native EventSource API.
 * Handles automatic lifecycle management, typed subscriptions, and payload parsing.
 */
export class TypedEventSource<TMap extends object> {
    private source: EventSource | null = null;
    private readonly listeners = new Map<keyof TMap, Set<SSEEventHandler<unknown>>>();
    private readonly errorListeners = new Set<SSEErrorHandler>();

    constructor(private readonly endpoint: string) { }

    public connect(): void {
        if (this.source) {
            return;
        }

        this.source = new EventSource(this.endpoint);

        this.source.onerror = (event: Event) => {
            this.errorListeners.forEach((listener) => listener(event));
        };

        for (const [eventName, handlers] of this.listeners.entries()) {
            this.bindNativeListener(eventName as string, handlers);
        }
    }

    public subscribe<K extends keyof TMap>(event: K, handler: SSEEventHandler<TMap[K]>): () => void {
        let handlers = this.listeners.get(event);
        if (!handlers) {
            handlers = new Set();
            this.listeners.set(event, handlers);
            if (this.source) {
                this.bindNativeListener(event as string, handlers);
            }
        }

        handlers.add(handler as SSEEventHandler<unknown>);

        return () => {
            handlers.delete(handler as SSEEventHandler<unknown>);
            if (handlers.size === 0) {
                this.listeners.delete(event);
            }
        };
    }

    public onError(handler: SSEErrorHandler): () => void {
        this.errorListeners.add(handler);
        return () => {
            this.errorListeners.delete(handler);
        };
    }

    public disconnect(): void {
        if (this.source) {
            this.source.close();
            this.source = null;
        }
    }

    private bindNativeListener(eventName: string, handlers: Set<SSEEventHandler<unknown>>): void {
        this.source?.addEventListener(eventName, (event: MessageEvent<string>) => {
            try {
                const parsed = JSON.parse(event.data) as unknown;
                handlers.forEach((handler) => handler(parsed));
            } catch {
                // Malformed payloads discarded to maintain stream integrity
            }
        });
    }
}