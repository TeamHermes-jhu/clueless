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
  EndTurn,
  ViewLocationsRequest,
  ViewHandRequest,
  ViewHandResponse,
  CharacterSelectRequest,
  CharacterSelectChoices,
<<<<<<< HEAD
  StartGame,
=======
>>>>>>> a7a81cda092dd21f3cb6b64a46bd0af2f58c573e
}

export enum EventProducer {
  SERVER,
  CLIENT,
}

export type AppEvent = ProtoAppEvent;

export type EventType = AppEvent['type'];

export type EventByType<T extends EventType> = Extract<AppEvent, { type: T }>;
export type EventPayload<T extends EventType> = EventByType<T>['payload'];
