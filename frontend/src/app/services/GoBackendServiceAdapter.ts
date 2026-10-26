import { Injectable, signal } from '@angular/core';
import type { DescMessage, MessageShape } from '@bufbuild/protobuf';
import { create, createRegistry, fromBinary, fromJsonString, toBinary, toJsonString } from '@bufbuild/protobuf';
import { anyPack, AnySchema, anyUnpack } from '@bufbuild/protobuf/wkt';
import { Rooms } from '../common/events/event-types/gen/cards/v1/enums_pb';
import type { AppEvent } from '../common/events/event-types/AppEvents';
import { EventProducer, MessageType } from '../common/events/event-types/AppEvents';
import {
  EnvelopeSchema,
  MoveSchema,
  PlayerSchema,
  JoinRequestSchema,
  SessionListSchema,
  SessionSchema,
} from '../common/events/event-types/gen/session/v1/session_pb';
import type { Envelope } from '../common/events/event-types/gen/session/v1/session_pb';
import {
  AccusationRequestSchema,
  AccusationResponseSchema,
  SuggestionRequestSchema,
  SuggestionResponseSchema,
} from '../common/events/event-types/gen/suggestion/v1/suggestion_pb';
import type { ProtoAppEvent } from '../common/events/event-types/ProtoEvents';
import { AbstractServiceAdapter } from './AbstractServiceAdapter';

// Every generated message the backend can send, paired with the event it is published as.
// When a new message is added to gen/, add its schema here and its type to MessageType.
const PROTO_MESSAGES: readonly (readonly [DescMessage, ProtoAppEvent['type']])[] = [
  [SessionSchema, MessageType.Session],
  [SessionListSchema, MessageType.SessionList],
  [JoinRequestSchema, MessageType.JoinRequest],
  [PlayerSchema, MessageType.Player],
];

const registry = createRegistry(...PROTO_MESSAGES.map(([schema]) => schema));
const eventTypes = new Map(PROTO_MESSAGES.map(([schema, type]) => [schema.typeName, type]));

@Injectable({ providedIn: 'root' })
export class GoBackendServiceAdapter extends AbstractServiceAdapter {
  private readonly currentPlayerId = signal<string | null>(null);
  readonly playerId = this.currentPlayerId.asReadonly();

  constructor() {
    super('localhost', 8000, '/ws');
  }

  joinSession(sessionId: string, username: string): void {
    this.currentPlayerId.set(username);
    this.connect({ session: sessionId, player: username});
  }

  sendEnvelope(payload: Envelope['payload']): void {
    this.sendMessage(toJsonString(EnvelopeSchema, create(EnvelopeSchema, { payload })))
  }

  protected override parse(raw: unknown): AppEvent | null {
    if (typeof raw !== 'string') {
        console.warn('GoBackendServiceAdapter ignored a non-text frame', raw);
        return null;
    }

    const envelope = fromJsonString(EnvelopeSchema, raw);

    switch (envelope.payload.case) {
      case 'move':
        return { type: 'move', producer: EventProducer.SERVER, payload: envelope.payload.value };
      case 'suggest':
        return { type: 'suggest', producer: EventProducer.SERVER, payload: envelope.payload.value };
      case 'accuse':
        return { type: 'accuse', producer: EventProducer.SERVER, payload: envelope.payload.value };
      case 'msg':
        return { type: 'msg', producer: EventProducer.SERVER, payload: envelope.payload.value };
      case 'error':
        return { type: 'error', producer: EventProducer.SERVER, payload: envelope.payload.value };
      default:
        console.warn('GoBackendServiceAdapter ignored a payloadless envelope', raw);
        return null;
    }
  }
}

