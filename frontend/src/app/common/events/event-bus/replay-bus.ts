// src/events/replay-event-bus.ts
import type { Observable} from 'rxjs';
import { ReplaySubject, filter, share, takeUntil } from 'rxjs';
import type { EventByType, EventPayload, EventType } from '../event-types/AppEvents';
import type { EventEnvelope, EventMetadata } from './event-envelope';

// Event bus that replays recent events to new subscribers
// Useful for late-joining components that need historical context
export class ReplayEventBus {
    private subject: ReplaySubject<EventEnvelope>;
    private destroy$ = new ReplaySubject<void>(1);
    private events$: Observable<EventEnvelope>;

    constructor(
        private source = 'default',
        private bufferSize = 100,
        private windowTime?: number  // Optional time window in ms
    ) {
        // ReplaySubject buffers events and replays them to new subscribers
        this.subject = windowTime
        ? new ReplaySubject<EventEnvelope>(bufferSize, windowTime)
        : new ReplaySubject<EventEnvelope>(bufferSize);

        this.events$ = this.subject.asObservable().pipe(
        takeUntil(this.destroy$),
        share()
        );
    }

    publish<T extends EventType>(
        type: T,
        payload: EventPayload<T>,
    ): void {
        const event = { type, payload } as EventByType<T>;

        const metadata: EventMetadata = {
            eventId: `${Date.now()}-${Math.random().toString(36).substring(2, 15)}`,
            timestamp: new Date(),
            source: this.source,
        };

        this.subject.next({ event, metadata });
    }

    subscribe<T extends EventType>(
        type: T
    ): Observable<EventEnvelope<EventByType<T>>> {
        return this.events$.pipe(
        filter((envelope): envelope is EventEnvelope<EventByType<T>> =>
            envelope.event.type === type
        )
        );
    }

    // Get buffer contents synchronously
    getBufferedEvents(): EventEnvelope[] {
        const events: EventEnvelope[] = [];
        // ReplaySubject emits buffered events synchronously to this temporary subscription
        this.subject.subscribe(event => events.push(event)).unsubscribe();
        return events;
    }

    destroy(): void {
        this.destroy$.next();
        this.destroy$.complete();
        this.subject.complete();
    }
}