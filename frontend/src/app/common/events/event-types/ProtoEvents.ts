import { EventProducer, MessageType } from './AppEvents';
import type { Envelope } from './gen/envelope/v1/envelope_pb';
import type {
  Error,
  JoinRequest,
  JoinResponse,
  Message,
  Move,
  Player,
  Session,
  SessionList,
} from './gen/session/v1/session_pb';
import type {
  AccusationRequest,
  AccusationResponse,
  SuggestionRequest,
  SuggestionResponse,
} from './gen/suggestion/v1/suggestion_pb';

// Maps every server MessageType to the generated protobuf message it carries.
export interface ProtoPayloadMap {
  [MessageType.Envelope]: Envelope;
  [MessageType.Player]: Player;
  [MessageType.Session]: Session;
  [MessageType.SessionList]: SessionList;
  [MessageType.JoinRequest]: JoinRequest;
  [MessageType.JoinResponse]: JoinResponse;
  [MessageType.Move]: Move;
  [MessageType.Message]: Message;
  [MessageType.Error]: Error;
  [MessageType.SuggestionRequest]: SuggestionRequest;
  [MessageType.SuggestionResponse]: SuggestionResponse;
  [MessageType.AccusationRequest]: AccusationRequest;
  [MessageType.AccusationResponse]: AccusationResponse;
}

export type ProtoMessageType = keyof ProtoPayloadMap;

// Events whose payload is a generated protobuf message received from the Go backend.
export type ProtoAppEvent = {
  [T in ProtoMessageType]: {
    type: T;
    producer: EventProducer.SERVER;
    payload: ProtoPayloadMap[T];
  };
}[ProtoMessageType];

type EnvelopeCase = NonNullable<Envelope['payload']['case']>;
type EnvelopeValue<K extends EnvelopeCase> = Extract<Envelope['payload'], { case: K }>['value'];

// For each case in Envelope oneof, map to correct to MessageType
type MessageTypeFor<P> = {
  [T in ProtoMessageType]: [ProtoPayloadMap[T]] extends [P]
    ? [P] extends [ProtoPayloadMap[T]]
      ? T
      : never
    : never;
}[ProtoMessageType];

export const ENVELOPE_MESSAGE_TYPES = {
  move: MessageType.Move,
  msg: MessageType.Message,
  suggest: MessageType.SuggestionRequest,
  accuse: MessageType.AccusationRequest,
  error: MessageType.Error,
} as const satisfies { [K in EnvelopeCase]: MessageTypeFor<EnvelopeValue<K>> };

// Converts a received Envelope into its bus event, or null if it has no payload.
export function envelopeToEvent(envelope: Envelope): ProtoAppEvent | null {
  if (envelope.payload.case === undefined) {
    return null;
  }

  return {
    type: ENVELOPE_MESSAGE_TYPES[envelope.payload.case],
    producer: EventProducer.SERVER,
    payload: envelope.payload.value,
  } as ProtoAppEvent;
}
