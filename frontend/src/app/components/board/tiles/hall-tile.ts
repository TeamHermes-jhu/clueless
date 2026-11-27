import { hallKey, type MoveTarget } from '../../../common/board/location';
import type { Halls } from '../../../common/events/event-types/gen/cards/v1/cards_pb';
import { Tile, type IconOffset } from './tile';

export type HallOrientation = 'horizontal' | 'vertical';

const HALL_IMAGES: Record<HallOrientation, string> = {
  horizontal: '/assets/board/hall-horizontal-repeat.png',
  vertical: '/assets/board/hall-vertical-repeat.png',
};

export class HallTile extends Tile {
  readonly kind = 'hall';

  constructor(
    readonly hall: Halls,
    name: string,
    readonly orientation: HallOrientation,
    row: number,
    col: number,
  ) {
    super(hallKey(hall), name, HALL_IMAGES[orientation], row, col);
  }

  override get cssClass(): string {
    return `hall ${this.orientation}`;
  }

  canEnter(occupants: readonly string[]): boolean {
    return occupants.length === 0;
  }

  toMoveTarget(): MoveTarget {
    return { case: 'hall', value: this.hall };
  }

  iconOffsets(count: number): readonly IconOffset[] {
    return Array.from({ length: count }, () => ({ x: 0, y: 0 }));
  }
}
