import type { Signal } from '@angular/core';
import { computed, effect, inject, Injectable, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { filter } from 'rxjs';
import { isLocationKey, locationKeyOf, type LocationKey } from '../board/location';
import { EventBus } from '../events/event-bus/event-bus';
import { EventProducer, MessageType } from '../events/event-types/AppEvents';
import { PlayerStore } from './PlayerStore';

interface BoardState {
  locations: Record<string, LocationKey>;
  currentPlayerId: string | null;
  movedThisTurn: boolean;
}

const EMPTY_STATE: BoardState = { locations: {}, currentPlayerId: null, movedThisTurn: false };

function isBoardState(value: unknown): value is BoardState {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const { locations, currentPlayerId, movedThisTurn } = value as Record<string, unknown>;
  return (
    typeof locations === 'object' &&
    locations !== null &&
    Object.values(locations).every(isLocationKey) &&
    (currentPlayerId === null || typeof currentPlayerId === 'string') &&
    typeof movedThisTurn === 'boolean'
  );
}

@Injectable({ providedIn: 'root' })
export class BoardStateStore {
  static readonly STORAGE_KEY = 'clueless.board';

  private readonly playerStore = inject(PlayerStore);
  private readonly state = signal<BoardState>(BoardStateStore.load());
  private readonly requestedMove = signal<LocationKey | null>(null);

  readonly locations: Signal<Readonly<Record<string, LocationKey>>> = computed(
    () => this.state().locations,
  );
  readonly currentPlayerId: Signal<string | null> = computed(() => this.state().currentPlayerId);
  readonly pendingMove: Signal<LocationKey | null> = this.requestedMove.asReadonly();
  readonly isMyTurn: Signal<boolean> = computed(() => {
    const me = this.playerStore.playerId();
    return me !== null && me === this.state().currentPlayerId;
  });
  readonly canMove: Signal<boolean> = computed(
    () => this.isMyTurn() && !this.state().movedThisTurn && this.requestedMove() === null,
  );

  constructor() {
    const eventBus = inject(EventBus);

    effect(() => BoardStateStore.save(this.state()));

    eventBus
      .on(MessageType.StartGame)
      .pipe(takeUntilDestroyed())
      .subscribe(() => {
        this.state.set(EMPTY_STATE);
        this.requestedMove.set(null);
      });

    eventBus
      .subscribe(MessageType.Move)
      .pipe(takeUntilDestroyed())
      .subscribe(({ event }) => {
        const destination = locationKeyOf(event.payload.moveTo);
        if (destination === null) {
          return;
        }

        if (event.producer === EventProducer.CLIENT) {
          this.requestedMove.set(destination);
          return;
        }

        this.applyServerMove(event.payload.playerId, destination);
      });

    eventBus
      .on(MessageType.TurnIndication)
      .pipe(takeUntilDestroyed())
      .subscribe(({ currentPlayer }) => {
        this.state.update((current) => ({
          ...current,
          currentPlayerId: currentPlayer?.playerName ?? null,
          movedThisTurn: false,
        }));
        this.requestedMove.set(null);
      });

    eventBus
      .subscribe(MessageType.Error)
      .pipe(
        filter(({ event }) => event.producer === EventProducer.SERVER),
        takeUntilDestroyed(),
      )
      .subscribe(() => this.requestedMove.set(null));
  }

  private applyServerMove(playerId: string, destination: LocationKey): void {
    const answersRequest =
      playerId === this.playerStore.playerId() && destination === this.requestedMove();

    this.state.update((current) => ({
      ...current,
      locations: { ...current.locations, [playerId]: destination },
      movedThisTurn: current.movedThisTurn || answersRequest,
    }));

    if (answersRequest) {
      this.requestedMove.set(null);
    }
  }

  private static load(): BoardState {
    try {
      const parsed: unknown = JSON.parse(
        sessionStorage.getItem(BoardStateStore.STORAGE_KEY) ?? 'null',
      );
      return isBoardState(parsed) ? parsed : EMPTY_STATE;
    } catch {
      return EMPTY_STATE;
    }
  }

  private static save(state: BoardState): void {
    try {
      sessionStorage.setItem(BoardStateStore.STORAGE_KEY, JSON.stringify(state));
    } catch (err) {
      console.warn(`Could not save board to "${BoardStateStore.STORAGE_KEY}":`, err);
    }
  }
}
