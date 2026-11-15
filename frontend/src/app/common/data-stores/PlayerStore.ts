import type { Signal } from '@angular/core';
import { computed, effect, Injectable, signal } from '@angular/core';
import type { Characters } from '../events/event-types/gen/cards/v1/cards_pb';

export interface PlayerData {
  playerId: string | null;
  character: Characters | null;
}

const EMPTY_PLAYER: PlayerData = { playerId: null, character: null };

function isPlayerData(value: unknown): value is PlayerData {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  const { playerId, character } = candidate;
  return (
    (playerId === null || typeof playerId === 'string') &&
    (character === null || typeof character === 'number')
  );
}

// Holds this client's own player
@Injectable()
export class PlayerStore {
  static readonly STORAGE_KEY = 'clueless.player';

  private readonly data = signal<PlayerData>(PlayerStore.load());

  readonly player: Signal<PlayerData> = this.data.asReadonly();
  readonly playerId: Signal<string | null> = computed(() => this.data().playerId);
  readonly character: Signal<Characters | null> = computed(() => this.data().character);

  constructor() {
    effect(() => PlayerStore.save(this.data()));
  }

  setPlayerId(playerId: string): void {
    this.data.update((current) =>
      current.playerId === playerId ? current : { playerId, character: null },
    );
  }

  setCharacter(character: Characters | null): void {
    this.data.update((current) => ({ ...current, character }));
  }

  clear(): void {
    this.data.set(EMPTY_PLAYER);
  }

  private static load(): PlayerData {
    try {
      const parsed: unknown = JSON.parse(sessionStorage.getItem(PlayerStore.STORAGE_KEY) ?? 'null');
      return isPlayerData(parsed) ? parsed : EMPTY_PLAYER;
    } catch {
      return EMPTY_PLAYER;
    }
  }

  private static save(data: PlayerData): void {
    try {
      sessionStorage.setItem(PlayerStore.STORAGE_KEY, JSON.stringify(data));
    } catch (err) {
      console.warn(`Could not save player to "${PlayerStore.STORAGE_KEY}":`, err);
    }
  }
}
