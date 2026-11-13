import { Injectable, signal } from '@angular/core';
import { create, fromJsonString, toJsonString } from '@bufbuild/protobuf';
import type { AppEvent } from '../common/events/event-types/AppEvents';
import { EventProducer, MessageType } from '../common/events/event-types/AppEvents';
import type { Envelope } from '../common/events/event-types/gen/envelope/v1/envelope_pb';
import { EnvelopeSchema } from '../common/events/event-types/gen/envelope/v1/envelope_pb';
import { SuggestionRequestSchema } from '../common/events/event-types/gen/suggestion/v1/suggestion_pb';
import { envelopeToEvent } from '../common/events/event-types/ProtoEvents';
import { AbstractServiceAdapter } from './AbstractServiceAdapter';

interface StoredSession {
  sessionId: string;
  username: string;
}

@Injectable({ providedIn: 'root' })
export class GoBackendServiceAdapter extends AbstractServiceAdapter {
  // Kept in sessionStorage so a page refresh can rejoin the same session.
  private static readonly SESSION_KEY = 'clueless.session';

  private readonly currentPlayerId = signal<string | null>(null);
  readonly playerId = this.currentPlayerId.asReadonly();

  constructor() {
    super('localhost', 8000, '/ws');
    this.registerCommandHandlers();

    const stored = this.loadSession();
    if (stored) {
      this.joinSession(stored.sessionId, stored.username);
    }
  }

  joinSession(sessionId: string, username: string): void {
    this.currentPlayerId.set(username);
    this.saveSession({ sessionId, username });
    this.connect({ session: sessionId, player: username });
  }

  sendEnvelope(payload: Envelope['payload']): void {
    this.sendMessage(toJsonString(EnvelopeSchema, create(EnvelopeSchema, { payload })));
  }

  // Forward client-produced events from the bus to the server, stamped with this player's id.
  // Server-produced events of the same types are ignored so they are not echoed back.
  private registerCommandHandlers(): void {
    this.handlers
      .on(
        MessageType.Move,
        (move) => {
          this.sendAsPlayer((playerId) => ({ case: 'move', value: { ...move, playerId } }));
        },
        EventProducer.CLIENT,
      )
      .on(
        MessageType.SuggestionRequest,
        (suggestion) => {
          this.sendAsPlayer((playerId) => ({
            case: 'suggest',
            value: { ...suggestion, playerId },
          }));
        },
        EventProducer.CLIENT,
      )
      .on(
        MessageType.AccusationRequest,
        (accusation) => {
          this.sendAsPlayer((playerId) => ({
            case: 'accuse',
            value: {
              ...accusation,
              suggestion: {
                ...(accusation.suggestion ?? create(SuggestionRequestSchema)),
                playerId,
              },
            },
          }));
        },
        EventProducer.CLIENT,
      )
      .on(
        MessageType.EndTurn,
        (endTurn) => {
          this.sendAsPlayer((playerId) => ({ case: 'endTurn', value: { ...endTurn, playerId } }));
        },
        EventProducer.CLIENT,
      )
      .on(
        MessageType.ViewHandRequest,
        (viewHand) => {
          this.sendAsPlayer((playerId) => ({ case: 'viewHand', value: { ...viewHand, playerId } }));
        },
        EventProducer.CLIENT,
      )
      .on(
        MessageType.ViewLocationsRequest,
        (viewLocations) => {
          this.sendAsPlayer((playerId) => ({
            case: 'viewLocations',
            value: { ...viewLocations, playerId },
          }));
        },
        EventProducer.CLIENT,
      )
      .on(
        MessageType.CharacterSelectRequest,
        (characterSelect) => {
          this.sendAsPlayer((playerId) => ({
            case: 'characterSelect',
            value: { ...characterSelect, playerId },
          }));
        },
        EventProducer.CLIENT,
      );
  }

  private sendAsPlayer(build: (playerId: string) => Envelope['payload']): void {
    const playerId = this.playerId();
    if (!playerId) {
      console.error('Cannot send a command before a session is joined.');
      return;
    }
    this.sendEnvelope(build(playerId));
  }

  // The stored session is likely gone (e.g. server restart), so stop trying to rejoin it.
  protected override onConnectionFailed(err: unknown): void {
    super.onConnectionFailed(err);
    this.currentPlayerId.set(null);
    this.clearSession();
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

  private loadSession(): StoredSession | null {
    try {
      const parsed: unknown = JSON.parse(
        sessionStorage.getItem(GoBackendServiceAdapter.SESSION_KEY) ?? 'null',
      );
      if (
        parsed !== null &&
        typeof parsed === 'object' &&
        typeof (parsed as StoredSession).sessionId === 'string' &&
        typeof (parsed as StoredSession).username === 'string'
      ) {
        return parsed as StoredSession;
      }
    } catch {
      // Fall through to treat unreadable storage as no session.
    }
    return null;
  }

  private saveSession(session: StoredSession): void {
    try {
      sessionStorage.setItem(GoBackendServiceAdapter.SESSION_KEY, JSON.stringify(session));
    } catch (err) {
      console.warn('Could not save session for reconnect:', err);
    }
  }

  private clearSession(): void {
    try {
      sessionStorage.removeItem(GoBackendServiceAdapter.SESSION_KEY);
    } catch {
      // Nothing to clear if storage is unavailable.
    }
  }
}
