import { inject, Injectable, OnDestroy } from "@angular/core";
import { Observable, retry, Subscription, throwError, timer } from 'rxjs';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { EventBus } from "../common/events/event-bus/event-bus";
import { EventHandlerRegistry } from "../common/events/event-bus/event-handler";
import { AppEvent } from "../common/events/event-types/AppEvents";

@Injectable()
export abstract class AbstractServiceAdapter implements OnDestroy {

    protected readonly eventBus = inject(EventBus);
    protected handlers: EventHandlerRegistry;
    private socket$: WebSocketSubject<unknown>;
    private inbound: Subscription;

    private static readonly MAX_CONNECT_ATTEMPTS = 3;
    private static readonly RETRY_DELAY_MS = 5000;
    private failedAttempts = 0;

    constructor(host: string, port: number, path: string) {
        const verifiedPath = path.startsWith('/') ? path : `/${path}`;
        this.socket$ = webSocket<unknown>({
            url: `ws://${host}:${port}${verifiedPath}`,
            deserializer: (e: MessageEvent) => e.data,
            openObserver: { next: () => { this.failedAttempts = 0; } },
        });
        this.handlers = new EventHandlerRegistry(this.eventBus);

        this.inbound = this.socket$.pipe(
            retry({ delay: (err) => this.scheduleReconnect(err) }),
        ).subscribe({
            next: (raw) => this.publish(raw),
            error: (err) => this.onConnectionFailed(err),
        });
    }

    // Called on connection fail. Override to handle failure further.
    protected onConnectionFailed(err: unknown): void {
        console.error(
            `${this.constructor.name} could not connect after ${AbstractServiceAdapter.MAX_CONNECT_ATTEMPTS} attempts; giving up.`,
            err,
        );
    }

    private scheduleReconnect(err: unknown): Observable<number> {
        this.failedAttempts++;
        if (this.failedAttempts >= AbstractServiceAdapter.MAX_CONNECT_ATTEMPTS) {
            return throwError(() => err);
        }
        console.warn(
            `${this.constructor.name} connection failed (attempt ${this.failedAttempts}/${AbstractServiceAdapter.MAX_CONNECT_ATTEMPTS}); retrying in ${AbstractServiceAdapter.RETRY_DELAY_MS / 1000}s.`,
        );
        return timer(AbstractServiceAdapter.RETRY_DELAY_MS);
    }

    // Convert one raw websocket frame into an event the bus understands.
    protected abstract parse(raw: unknown): AppEvent | null;

    sendMessage(message: unknown) {
        this.socket$.next(message);
    }

    closeConnection() {
        this.inbound.unsubscribe();
        this.socket$.complete();
    }

    ngOnDestroy(): void {
        this.handlers.clear();
        this.closeConnection();
    }

    private publish(raw: unknown): void {
        let event: AppEvent | null;
        try {
            event = this.parse(raw);
        } catch (err) {
            console.error(`${this.constructor.name} failed to parse message:`, raw, err);
            return;
        }
        if (event) {
            this.eventBus.publish(event.type, event.payload);
        }
    }
}
