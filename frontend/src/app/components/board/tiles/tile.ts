import type { LocationKey, MoveTarget } from '../../../common/board/location';

export interface IconOffset {
  x: number;
  y: number;
}

export abstract class Tile {
  abstract readonly kind: 'room' | 'hall';

  protected constructor(
    readonly key: LocationKey,
    readonly name: string,
    readonly image: string,
    readonly row: number,
    readonly col: number,
  ) {}

  get cssClass(): string {
    return this.kind;
  }

  abstract canEnter(occupants: readonly string[]): boolean;

  abstract toMoveTarget(): MoveTarget;

  abstract iconOffsets(count: number): readonly IconOffset[];
}
