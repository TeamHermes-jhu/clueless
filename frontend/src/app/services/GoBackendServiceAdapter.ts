import { Injectable, signal } from '@angular/core';
import { create, fromJsonString, toJsonString } from '@bufbuild/protobuf';
import type { AppEvent } from '../common/events/event-types/AppEvents';
import type { Envelope } from '../common/events/event-types/gen/envelope/v1/envelope_pb';
import { EnvelopeSchema } from '../common/events/event-types/gen/envelope/v1/envelope_pb';
import { envelopeToEvent } from '../common/events/event-types/ProtoEvents';
import { AbstractServiceAdapter } from './AbstractServiceAdapter';

@Injectable({ providedIn: 'root' })
export class GoBackendServiceAdapter extends AbstractServiceAdapter {
  private readonly currentPlayerId = signal<string | null>(null);
  readonly playerId = this.currentPlayerId.asReadonly();

  constructor() {
    super('localhost', 8000, '/ws');
  }

  joinSession(sessionId: string, username: string): void {
    this.currentPlayerId.set(username);
    this.connect({ session: sessionId, player: username });
  }

  sendEnvelope(payload: Envelope['payload']): void {
    this.sendMessage(toJsonString(EnvelopeSchema, create(EnvelopeSchema, { payload })));
  }

  protected override parse(raw: unknown): AppEvent | null {
    if (typeof raw !== 'string') {
      console.warn('GoBackendServiceAdapter ignored a non-text frame', raw);
      return null;
    }

    const event = envelopeToEvent(fromJsonString(EnvelopeSchema, raw));
    if (event === null) {
      console.warn('GoBackendServiceAdapter ignored a payloadless envelope', raw);
    }
    return event;
  }
}
