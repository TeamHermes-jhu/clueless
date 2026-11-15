import type { Signal } from '@angular/core';
import { computed, inject, Injectable } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { EMPTY } from 'rxjs';
import { EventBus } from '../events/event-bus/event-bus';
import { MessageType } from '../events/event-types/AppEvents';
import type { Characters } from '../events/event-types/gen/cards/v1/cards_pb';
import { AbstractStore } from './AbstractStore';

export interface CharacterAssignment {
  playerId: string;
  character: Characters;
}

function isCharacterAssignment(value: unknown): value is CharacterAssignment {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  return typeof candidate['playerId'] === 'string' && typeof candidate['character'] === 'number';
}

@Injectable({ providedIn: 'root' })
export class CharacterSelectionStore extends AbstractStore<CharacterAssignment> {
  static readonly STORAGE_KEY = 'clueless.characterSelections';

  constructor() {
    const eventBus = inject(EventBus);

    super(EMPTY, {
      key: CharacterSelectionStore.STORAGE_KEY,
      isItem: isCharacterAssignment,
    });

    eventBus
      .on(MessageType.CharacterSelectChoices)
      .pipe(takeUntilDestroyed())
      .subscribe(({ assignments }) => {
        this.items.set(assignments.map(({ playerId, character }) => ({ playerId, character })));
      });
  }

  playerIdFor(character: Characters): Signal<string | undefined> {
    return computed(() => this.items().find((a) => a.character === character)?.playerId);
  }

  characterFor(playerId: string): Signal<Characters | undefined> {
    return computed(() => this.items().find((a) => a.playerId === playerId)?.character);
  }
}
