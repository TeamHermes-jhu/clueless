import { inject, Injectable } from '@angular/core';
import { map } from 'rxjs';
import { EventBus } from '../events/event-bus/event-bus';
import { MessageType } from '../events/event-types/AppEvents';
import { AbstractStore } from './AbstractStore';

@Injectable({ providedIn: 'root' })
export class TextMessageStore extends AbstractStore<string> {
  static readonly STORAGE_KEY = 'clueless.textMessages';

  constructor() {
    const eventBus = inject(EventBus);

    super(eventBus.on(MessageType.Message).pipe(map((payload) => payload.msg)), {
      key: TextMessageStore.STORAGE_KEY,
      isItem: (value): value is string => typeof value === 'string',
    });
  }
}
