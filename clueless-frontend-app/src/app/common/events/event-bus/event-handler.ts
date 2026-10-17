// src/events/event-handler.ts
import { Subscription } from 'rxjs';
import { EventByType, EventPayload, EventType } from '../event-types/AppEvents';
import { EventBus } from './event-bus';
import { EventEnvelope } from './event-envelope';

// Handler function signature
type EventHandler<T extends EventType> = (
    payload: EventPayload<T>,
    metadata: EventEnvelope<EventByType<T>>['metadata']
) => void | Promise<void>;

// Registration for cleanup
interface HandlerRegistration {
    eventType: EventType;
    handler: EventHandler<any>;
    subscription: Subscription;
}

// Event handler registry with automatic subscription management
// Provides a cleaner API for registering multiple handlers
export class EventHandlerRegistry {
    private handlers: HandlerRegistration[] = [];

    constructor(private eventBus: EventBus) {}

    // Register a handler for a specific event type
    on<T extends EventType>(
        type: T,
        handler: EventHandler<T>
    ): this {
        const subscription = this.eventBus.subscribe(type).subscribe({
        next: async (envelope) => {
            try {
            await handler(envelope.event.payload, envelope.metadata);
            } catch (error) {
            console.error(`Error in handler for ${type}:`, error);
            }
        },
        });

        this.handlers.push({
        eventType: type,
        handler,
        subscription,
        });

        return this;
    }

    // Register handlers for multiple events at once
    onMany<T extends EventType>(
        types: T[],
        handler: EventHandler<T>
    ): this {
        for (const type of types) {
        this.on(type, handler);
        }
        return this;
    }

    // Unregister all handlers for a specific event type
    off(type: EventType): void {
        const handlersToRemove = this.handlers.filter(h => h.eventType === type);

        for (const handler of handlersToRemove) {
        handler.subscription.unsubscribe();
        }

        this.handlers = this.handlers.filter(h => h.eventType !== type);
    }

    // Unregister all handlers
    clear(): void {
        for (const handler of this.handlers) {
        handler.subscription.unsubscribe();
        }
        this.handlers = [];
    }
}