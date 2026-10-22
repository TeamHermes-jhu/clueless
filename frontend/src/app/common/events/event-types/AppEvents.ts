import type { TextMessageRecieve, TextMessageSend } from './ClientEvents';
import type { ProtoAppEvent } from './ProtoEvents';

// Must include EVERY event type here
export enum MessageType {
  TextMessageRecieve,
  TextMessageSend,

  // Protobuf messages from gen/session
  Session,
  SessionList,
  SessionJoinRequest,
  Player,

  // Protobuf messages from gen/suggestion
  SuggestionMessage,
  AccusationMessage,
  SuggestionResponseMessage,
  AccusationResponseMessage,
}

export enum EventProducer {
  SERVER,
  CLIENT,
}

export type AppEvent = TextMessageRecieve | TextMessageSend | ProtoAppEvent;

export type EventType = AppEvent['type'];

export type EventByType<T extends EventType> = Extract<AppEvent, { type: T }>;
export type EventPayload<T extends EventType> = EventByType<T>['payload'];
