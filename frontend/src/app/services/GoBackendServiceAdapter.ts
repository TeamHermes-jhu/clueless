import { Injectable, signal } from '@angular/core';
import { create, fromJsonString, toJsonString } from '@bufbuild/protobuf';
import type { AppEvent } from '../common/events/event-types/AppEvents';
import type { Envelope } from '../common/events/event-types/gen/envelope/v1/envelope_pb';
import { EnvelopeSchema } from '../common/events/event-types/gen/envelope/v1/envelope_pb';
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
