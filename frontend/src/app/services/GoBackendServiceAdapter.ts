import { Injectable } from '@angular/core';
import type { DescMessage, MessageShape } from '@bufbuild/protobuf';
import { createRegistry, fromBinary, fromJsonString, toBinary } from '@bufbuild/protobuf';
import { anyPack, anyUnpack, AnySchema } from '@bufbuild/protobuf/wkt';
import type { AppEvent } from '../common/events/event-types/AppEvents';
import { EventProducer, MessageType } from '../common/events/event-types/AppEvents';
import {
  PlayerSchema,
  SessionJoinRequestSchema,
  SessionListSchema,
  SessionSchema,
} from '../common/events/event-types/gen/session/v1/session_messages_pb';
import {
  AccusationMessageSchema,
  AccusationResponseMessageSchema,
  SuggestionMessageSchema,
  SuggestionResponseMessageSchema,
} from '../common/events/event-types/gen/suggestion/v1/suggest_accuse_messages_pb';
import type { ProtoAppEvent } from '../common/events/event-types/ProtoEvents';
import { AbstractServiceAdapter } from './AbstractServiceAdapter';

// Every generated message the backend can send, paired with the event it is published as.
// When a new message is added to gen/, add its schema here and its type to MessageType.
const PROTO_MESSAGES: readonly (readonly [DescMessage, ProtoAppEvent['type']])[] = [
  [SessionSchema, MessageType.Session],
  [SessionListSchema, MessageType.SessionList],
  [SessionJoinRequestSchema, MessageType.SessionJoinRequest],
  [PlayerSchema, MessageType.Player],
  [SuggestionMessageSchema, MessageType.SuggestionMessage],
  [AccusationMessageSchema, MessageType.AccusationMessage],
  [SuggestionResponseMessageSchema, MessageType.SuggestionResponseMessage],
  [AccusationResponseMessageSchema, MessageType.AccusationResponseMessage],
];

const registry = createRegistry(...PROTO_MESSAGES.map(([schema]) => schema));
const eventTypes = new Map(PROTO_MESSAGES.map(([schema, type]) => [schema.typeName, type]));

// Talks to the Go backend. Every frame is a google.protobuf.Any wrapping one generated
// message: binary frames hold the encoded Any, and text frames hold its JSON form
// (e.g. {"@type": "type.googleapis.com/session.v1.Session", ...}), which is handy for testing.
@Injectable({ providedIn: 'root' })
export class GoBackendServiceAdapter extends AbstractServiceAdapter {
  constructor() {
    super('localhost', 8080, '/ws');
  }

  // Wrap a generated message in an Any and send it as a binary frame.
  sendProto<Desc extends DescMessage>(schema: Desc, message: MessageShape<Desc>): void {
    this.sendMessage(toBinary(AnySchema, anyPack(schema, message)));
  }

  protected override parse(raw: unknown): AppEvent | null {
    let any;
    if (raw instanceof ArrayBuffer) {
      any = fromBinary(AnySchema, new Uint8Array(raw));
    } else if (typeof raw === 'string') {
      any = fromJsonString(AnySchema, raw, { registry });
    } else {
      console.warn('GoBackendServiceAdapter ignored a frame that is neither text nor binary:', raw);
      return null;
    }

    const message = anyUnpack(any, registry);
    const type = message && eventTypes.get(message.$typeName);
    if (!message || type === undefined) {
      console.warn('GoBackendServiceAdapter ignored unknown message type:', any.typeUrl);
      return null;
    }
    return { type, producer: EventProducer.SERVER, payload: message } as ProtoAppEvent;
  }
}
