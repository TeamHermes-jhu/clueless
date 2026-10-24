import type { Observable } from 'rxjs';
import { Subject, filter, map, share, takeUntil } from 'rxjs';
import type { EventByType, EventPayload, EventType } from '../event-types/AppEvents';
import type { EventEnvelope, EventMetadata } from './event-envelope';

export class EventBus {
  private subject = new Subject<EventEnvelope>();
  private destroy$ = new Subject<void>();

  // Shared observable for all events
  private events$ = this.subject.asObservable().pipe(takeUntil(this.destroy$), share());

  constructor(private source = 'default') {}

  // Publish an event with automatic metadata generation
  publish<T extends EventType>(type: T, payload: EventPayload<T>): void {
    const event = { type, payload } as EventByType<T>;

    const metadata: EventMetadata = {
      eventId: this.generateEventId(),
      timestamp: new Date(),
      source: this.source,
    };

    this.subject.next({ event, metadata });
  }

  // Subscribe to all events
  subscribeAll(): Observable<EventEnvelope> {
    return this.events$;
  }

  // Subscribe to a specific event type with full type safety
  subscribe<T extends EventType>(type: T): Observable<EventEnvelope<EventByType<T>>> {
    return this.events$.pipe(
      filter((envelope): envelope is EventEnvelope<EventByType<T>> => envelope.event.type === type),
    );
  }

  // Subscribe to multiple event types
  subscribeMany<T extends EventType>(types: T[]): Observable<EventEnvelope<EventByType<T>>> {
    return this.events$.pipe(
      filter((envelope): envelope is EventEnvelope<EventByType<T>> =>
        types.includes(envelope.event.type as T),
      ),
    );
  }

  // Subscribe with a custom filter
  subscribeWhere(predicate: (envelope: EventEnvelope) => boolean): Observable<EventEnvelope> {
    return this.events$.pipe(filter(predicate));
  }

  // Get just the payload for a specific event type
  on<T extends EventType>(type: T): Observable<EventPayload<T>> {
    return this.subscribe(type).pipe(map((envelope) => envelope.event.payload as EventPayload<T>));
  }

  // Cleanup subscriptions
  destroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.subject.complete();
  }

  private generateEventId(): string {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 15)}`;
  }
}
