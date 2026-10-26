import type { EventProducer, MessageType } from './AppEvents';
import type {
  Player,
  Session,
  JoinRequest,
  SessionList,
  Move,
  Message,
  Error,
} from './gen/session/v1/session_pb';
import type {
  AccusationRequest,
  AccusationResponse,
  SuggestionRequest,
  SuggestionResponse,
} from './gen/suggestion/v1/suggestion_pb';

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

export interface JoinRequestEvent {
  type: MessageType.JoinRequest;
  producer: EventProducer.SERVER;
  payload: JoinRequest;
}

export interface PlayerEvent {
  type: MessageType.Player;
  producer: EventProducer.SERVER;
  payload: Player;
}

export interface SuggestionEvent {
  type: 'suggest',
  producer: EventProducer.SERVER;
  payload: SuggestionRequest;
}

export interface AccusationEvent {
  type: 'accuse',
  producer: EventProducer.SERVER;
  payload: AccusationRequest;
}

export interface SuggestionResponseEvent {
  type: MessageType.SuggestionResponseMessage;
  producer: EventProducer.SERVER;
  payload: SuggestionResponse;
}

export interface AccusationResponseEvent {
  type: MessageType.AccusationResponseMessage;
  producer: EventProducer.SERVER;
  payload: AccusationResponse;
}

export interface MoveResponseEvent {
  type: 'move';
  producer: EventProducer.SERVER;
  payload: Move;
}

export interface MessageEvent {
  type: 'msg';
  producer: EventProducer.SERVER;
  payload: Message;
}

export interface ErrorEvent {
  type: 'error';
  producer: EventProducer.SERVER;
  payload: Error;
}

export type ProtoAppEvent =
  | SessionEvent
  | SessionListEvent
  | JoinRequestEvent
  | PlayerEvent
  | SuggestionEvent
  | AccusationEvent
  | SuggestionResponseEvent
  | AccusationResponseEvent
  | MoveResponseEvent
  | MessageEvent
  | ErrorEvent;

