import type { TextMessageRecieve, TextMessageSend } from './ClientEvents';

// Must include EVERY event type here
export enum MessageType {
  TextMessageRecieve,
  TextMessageSend,
}

export enum EventProducer {
  SERVER,
  CLIENT,
}

export type AppEvent = TextMessageRecieve | TextMessageSend;

export type EventType = AppEvent['type'];

export type EventByType<T extends EventType> = Extract<AppEvent, { type: T }>;
export type EventPayload<T extends EventType> = EventByType<T>['payload'];
