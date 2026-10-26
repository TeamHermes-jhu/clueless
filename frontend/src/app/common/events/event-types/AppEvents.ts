import type { ProtoAppEvent } from './ProtoEvents';

// Must include EVERY event type here
export enum MessageType {
  // List all client only messages here

  // List all server messages here
  Envelope,
  Player,
  Session,
  SessionList,
  JoinRequest,
  JoinResponse,
  Move,
  Message,
  Error,
  SuggestionRequest,
  SuggestionResponse,
  AccusationRequest,
  AccusationResponse,
}

export enum EventProducer {
  SERVER,
  CLIENT,
}

export type AppEvent = ProtoAppEvent;

export type EventType = AppEvent['type'];

export type EventByType<T extends EventType> = Extract<AppEvent, { type: T }>;
export type EventPayload<T extends EventType> = EventByType<T>['payload'];
