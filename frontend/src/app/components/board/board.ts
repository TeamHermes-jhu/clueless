import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { create } from '@bufbuild/protobuf';
import type { LocationKey } from '../../common/board/location';
import { BoardStateStore } from '../../common/data-stores/BoardStateStore';
import { CharacterSelectionStore } from '../../common/data-stores/CharacterSelectionStore';
import { PlayerStore } from '../../common/data-stores/PlayerStore';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import { MoveSchema } from '../../common/events/event-types/gen/session/v1/session_pb';
import { PlayerIconAccessService } from '../../services/PlayerIconAccessService';
import {
  adjacentTiles,
  BOARD_TILES,
  BOARD_TRACK_TOTAL,
  BOARD_TRACKS,
  tileAt,
  trackCenterPercent,
} from './board-layout';
import type { Tile } from './tiles/tile';

const ICON_SIZE_PERCENT = 5.5;

interface PlacedIcon {
  playerId: string;
  name: string;
  src: string | undefined;
  color: string;
  isMe: boolean;
  left: number;
  top: number;
}

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: 'app-board',
  styleUrl: './board.css',
  templateUrl: './board.html',
})
export class Board {
  private readonly eventBus = inject(EventBus);
  private readonly boardState = inject(BoardStateStore);
  private readonly playerStore = inject(PlayerStore);
  private readonly characterSelections = inject(CharacterSelectionStore);
  private readonly iconAccess = inject(PlayerIconAccessService);

  protected readonly tiles = BOARD_TILES;
  protected readonly tracks = BOARD_TRACKS;
  protected readonly trackTotal = BOARD_TRACK_TOTAL;
  protected readonly iconSize = ICON_SIZE_PERCENT;
  protected readonly pendingMove = this.boardState.pendingMove;

  private readonly occupants = computed(() => {
    const byTile = new Map<LocationKey, string[]>();
    const characters = this.characterSelections.getAll()();
    const characterOf = (id: string) =>
      characters.find((a) => a.playerId === id)?.character ?? Number.MAX_SAFE_INTEGER;

    for (const [playerId, key] of Object.entries(this.boardState.locations())) {
      byTile.set(key, [...(byTile.get(key) ?? []), playerId]);
    }
    for (const players of byTile.values()) {
      players.sort((a, b) => characterOf(a) - characterOf(b));
    }
    return byTile;
  });

  protected readonly available = computed(() => {
    const me = this.playerStore.playerId();
    const here = me === null ? undefined : this.boardState.locations()[me];
    if (!this.boardState.canMove() || here === undefined) {
      return new Set<LocationKey>();
    }

    const occupants = this.occupants();
    return new Set(
      adjacentTiles(here)
        .filter((tile) => tile.canEnter((occupants.get(tile.key) ?? []).filter((id) => id !== me)))
        .map((tile) => tile.key),
    );
  });

  protected readonly icons = computed<PlacedIcon[]>(() => {
    const me = this.playerStore.playerId();
    const characters = this.characterSelections.getAll()();
    const placed: PlacedIcon[] = [];

    for (const [key, players] of this.occupants()) {
      const tile = tileAt(key);
      if (!tile) {
        continue;
      }

      const offsets = tile.iconOffsets(players.length);
      players.forEach((playerId, i) => {
        const character = characters.find((a) => a.playerId === playerId)?.character;
        if (character === undefined) {
          return;
        }
        placed.push({
          playerId,
          name: `${this.iconAccess.getName(character)} (${playerId})`,
          src: this.iconAccess.getIcon(character)(),
          color: this.iconAccess.getColor(character),
          isMe: playerId === me,
          left: trackCenterPercent(tile.col) + offsets[i].x * ICON_SIZE_PERCENT,
          top: trackCenterPercent(tile.row) + offsets[i].y * ICON_SIZE_PERCENT,
        });
      });
    }
    return placed;
  });

  protected moveTo(tile: Tile): void {
    if (!this.available().has(tile.key)) {
      return;
    }

    this.eventBus.publish({
      type: MessageType.Move,
      producer: EventProducer.CLIENT,
      payload: create(MoveSchema, { moveTo: tile.toMoveTarget() }),
    });
  }
}
