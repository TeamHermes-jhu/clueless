import type { EventProducer, MessageType } from './AppEvents';
import type {
  Player,
  Session,
  SessionJoinRequest,
  SessionList,
} from './gen/session/v1/session_messages_pb';
import type {
  AccusationMessage,
  AccusationResponseMessage,
  SuggestionMessage,
  SuggestionResponseMessage,
} from './gen/suggestion/v1/suggest_accuse_messages_pb';

// Events whose payload is a generated protobuf message received from the Go backend.
export interface SessionEvent {
  type: MessageType.Session;
  producer: EventProducer.SERVER;
  payload: Session;
}

export interface SessionListEvent {
  type: MessageType.SessionList;
  producer: EventProducer.SERVER;
  payload: SessionList;
}

export interface SessionJoinRequestEvent {
  type: MessageType.SessionJoinRequest;
  producer: EventProducer.SERVER;
  payload: SessionJoinRequest;
}

export interface PlayerEvent {
  type: MessageType.Player;
  producer: EventProducer.SERVER;
  payload: Player;
}

export interface SuggestionEvent {
  type: MessageType.SuggestionMessage;
  producer: EventProducer.SERVER;
  payload: SuggestionMessage;
}

export interface AccusationEvent {
  type: MessageType.AccusationMessage;
  producer: EventProducer.SERVER;
  payload: AccusationMessage;
}

export interface SuggestionResponseEvent {
  type: MessageType.SuggestionResponseMessage;
  producer: EventProducer.SERVER;
  payload: SuggestionResponseMessage;
}

export interface AccusationResponseEvent {
  type: MessageType.AccusationResponseMessage;
  producer: EventProducer.SERVER;
  payload: AccusationResponseMessage;
}

export type ProtoAppEvent =
  | SessionEvent
  | SessionListEvent
  | SessionJoinRequestEvent
  | PlayerEvent
  | SuggestionEvent
  | AccusationEvent
  | SuggestionResponseEvent
  | AccusationResponseEvent;
