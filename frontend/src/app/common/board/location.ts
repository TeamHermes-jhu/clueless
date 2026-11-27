import type { Halls, Rooms } from '../events/event-types/gen/cards/v1/cards_pb';
import type { Move } from '../events/event-types/gen/session/v1/session_pb';

export type LocationKey = `room:${number}` | `hall:${number}`;

export type MoveTarget = Move['moveTo'];

export function roomKey(room: Rooms): LocationKey {
  return `room:${room}`;
}

export function hallKey(hall: Halls): LocationKey {
  return `hall:${hall}`;
}

export function locationKeyOf(moveTo: MoveTarget): LocationKey | null {
  switch (moveTo.case) {
    case 'room':
      return roomKey(moveTo.value);
    case 'hall':
      return hallKey(moveTo.value);
    default:
      return null;
  }
}

export function isLocationKey(value: unknown): value is LocationKey {
  return typeof value === 'string' && /^(room|hall):\d+$/.test(value);
}
