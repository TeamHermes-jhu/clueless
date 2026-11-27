import type { LocationKey } from '../../common/board/location';
import { Halls, Rooms } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { HallTile } from './tiles/hall-tile';
import { RoomTile } from './tiles/room-tile';
import type { Tile } from './tiles/tile';

const TRACKS = [3, 2, 3, 2, 3] as const;
const TRACK_TOTAL = TRACKS.reduce((sum, size) => sum + size, 0);

export const BOARD_TRACK_TOTAL = TRACK_TOTAL;

export const BOARD_TRACKS = TRACKS.map((size) => `${size}fr`).join(' ');

export function trackCenterPercent(index: number): number {
  const before = TRACKS.slice(0, index).reduce((sum, size) => sum + size, 0);
  return ((before + TRACKS[index] / 2) / TRACK_TOTAL) * 100;
}

function room(value: Rooms, name: string, image: string, row: number, col: number): RoomTile {
  return new RoomTile(value, name, `/assets/board/${image}.png`, row, col);
}

export const BOARD_TILES: readonly Tile[] = [
  room(Rooms.STUDY, 'Study', 'study', 0, 0),
  new HallTile(Halls.ONE, 'Hallway between Study and Hall', 'horizontal', 0, 1),
  room(Rooms.HALL, 'Hall', 'hall', 0, 2),
  new HallTile(Halls.TWO, 'Hallway between Hall and Lounge', 'horizontal', 0, 3),
  room(Rooms.LOUNGE, 'Lounge', 'lounge', 0, 4),

  new HallTile(Halls.SEVEN, 'Hallway between Study and Library', 'vertical', 1, 0),
  new HallTile(Halls.NINE, 'Hallway between Hall and Billiard Room', 'vertical', 1, 2),
  new HallTile(Halls.ELEVEN, 'Hallway between Lounge and Dining Room', 'vertical', 1, 4),

  room(Rooms.LIBRARY, 'Library', 'library', 2, 0),
  new HallTile(Halls.THREE, 'Hallway between Library and Billiard Room', 'horizontal', 2, 1),
  room(Rooms.BILLIARD_ROOM, 'Billiard Room', 'billiard-room', 2, 2),
  new HallTile(Halls.FOUR, 'Hallway between Billiard Room and Dining Room', 'horizontal', 2, 3),
  room(Rooms.DINING_ROOM, 'Dining Room', 'dining-room', 2, 4),

  new HallTile(Halls.EIGHT, 'Hallway between Library and Conservatory', 'vertical', 3, 0),
  new HallTile(Halls.TEN, 'Hallway between Billiard Room and Ballroom', 'vertical', 3, 2),
  new HallTile(Halls.TWELVE, 'Hallway between Dining Room and Kitchen', 'vertical', 3, 4),

  room(Rooms.CONSERVATORY, 'Conservatory', 'conservatory', 4, 0),
  new HallTile(Halls.FIVE, 'Hallway between Conservatory and Ballroom', 'horizontal', 4, 1),
  room(Rooms.BALLROOM, 'Ballroom', 'ballroom', 4, 2),
  new HallTile(Halls.SIX, 'Hallway between Ballroom and Kitchen', 'horizontal', 4, 3),
  room(Rooms.KITCHEN, 'Kitchen', 'kitchen', 4, 4),
];

const TILES_BY_KEY = new Map<LocationKey, Tile>(BOARD_TILES.map((tile) => [tile.key, tile]));

const SECRET_PASSAGES: readonly (readonly [Rooms, Rooms])[] = [
  [Rooms.STUDY, Rooms.KITCHEN],
  [Rooms.LOUNGE, Rooms.CONSERVATORY],
];

const ADJACENCY: ReadonlyMap<LocationKey, readonly Tile[]> = (() => {
  const adjacency = new Map<LocationKey, Tile[]>(BOARD_TILES.map((tile) => [tile.key, []]));
  const link = (a: Tile, b: Tile) => {
    adjacency.get(a.key)?.push(b);
    adjacency.get(b.key)?.push(a);
  };

  BOARD_TILES.forEach((a, i) => {
    for (const b of BOARD_TILES.slice(i + 1)) {
      if (Math.abs(a.row - b.row) + Math.abs(a.col - b.col) === 1) {
        link(a, b);
      }
    }
  });

  for (const [from, to] of SECRET_PASSAGES) {
    const a = TILES_BY_KEY.get(`room:${from}`);
    const b = TILES_BY_KEY.get(`room:${to}`);
    if (a && b) {
      link(a, b);
    }
  }

  return adjacency;
})();

export function tileAt(key: LocationKey): Tile | undefined {
  return TILES_BY_KEY.get(key);
}

export function adjacentTiles(key: LocationKey): readonly Tile[] {
  return ADJACENCY.get(key) ?? [];
}
