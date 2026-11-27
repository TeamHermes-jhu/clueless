import { roomKey, type MoveTarget } from '../../../common/board/location';
import type { Rooms } from '../../../common/events/event-types/gen/cards/v1/cards_pb';
import { Tile, type IconOffset } from './tile';

const SPACING = 1.1;

const CLUSTERS: readonly (readonly IconOffset[])[] = [
  [],
  [{ x: 0, y: 0 }],
  [
    { x: -0.5, y: 0 },
    { x: 0.5, y: 0 },
  ],
  [
    { x: 0, y: -0.45 },
    { x: -0.5, y: 0.45 },
    { x: 0.5, y: 0.45 },
  ],
  [
    { x: -0.5, y: -0.5 },
    { x: 0.5, y: -0.5 },
    { x: -0.5, y: 0.5 },
    { x: 0.5, y: 0.5 },
  ],
  [
    { x: -1, y: -0.5 },
    { x: 0, y: -0.5 },
    { x: 1, y: -0.5 },
    { x: -0.5, y: 0.5 },
    { x: 0.5, y: 0.5 },
  ],
  [
    { x: -1, y: -0.5 },
    { x: 0, y: -0.5 },
    { x: 1, y: -0.5 },
    { x: -1, y: 0.5 },
    { x: 0, y: 0.5 },
    { x: 1, y: 0.5 },
  ],
];

export class RoomTile extends Tile {
  readonly kind = 'room';

  constructor(
    readonly room: Rooms,
    name: string,
    image: string,
    row: number,
    col: number,
  ) {
    super(roomKey(room), name, image, row, col);
  }

  canEnter(): boolean {
    return true;
  }

  toMoveTarget(): MoveTarget {
    return { case: 'room', value: this.room };
  }

  iconOffsets(count: number): readonly IconOffset[] {
    const cluster = CLUSTERS[Math.min(count, CLUSTERS.length - 1)];
    const offsets = cluster.map(({ x, y }) => ({ x: x * SPACING, y: y * SPACING }));

    while (offsets.length < count) {
      offsets.push({ x: 0, y: 0 });
    }
    return offsets;
  }
}
