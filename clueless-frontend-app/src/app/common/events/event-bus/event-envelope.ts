import type { AppEvent } from '../event-types/AppEvents';

// Metadata attached to every event
export interface EventMetadata {
  eventId: string;
  timestamp: Date;
  source: string;
}

export interface EventEnvelope<T extends AppEvent = AppEvent> {
  metadata: EventMetadata;
  event: T;
}

export function createEventId(): string {
  return globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2, 12)}`;
}
