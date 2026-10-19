import { effect, inject, Injectable, signal } from '@angular/core';
import { EventBus } from '../events/event-bus/event-bus';
import { MessageType } from '../events/event-types/AppEvents';

const STORAGE_KEY = 'clueless.textMessages';

// Keeps received text messages across navigation, and saves them to
// sessionStorage so they survive page refreshes. sessionStorage is cleared when
// the tab or browser is closed, so the store starts empty on the next launch.
@Injectable({ providedIn: 'root' })
export class TextMessageStore {
  private readonly eventBus = inject(EventBus);
  private readonly messageList = signal<string[]>(loadMessages());

  readonly messages = this.messageList.asReadonly();

  constructor() {
    this.eventBus
      .on(MessageType.TextMessageRecieve)
      .subscribe((payload) => this.messageList.update((all) => [...all, payload.message]));
    effect(() => saveMessages(this.messageList()));
  }

  clear(): void {
    this.messageList.set([]);
  }
}

// Storage can be unavailable (private browsing, blocked site data) or hold bad data,
// so failures fall back to an empty list instead of breaking the app.
function loadMessages(): string[] {
  try {
    const parsed: unknown = JSON.parse(sessionStorage.getItem(STORAGE_KEY) ?? '[]');
    return Array.isArray(parsed) ? parsed.filter((m): m is string => typeof m === 'string') : [];
  } catch {
    return [];
  }
}

function saveMessages(messages: string[]): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(messages));
  } catch (err) {
    console.warn('TextMessageStore could not save messages:', err);
  }
}
