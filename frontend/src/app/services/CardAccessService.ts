import { Injectable } from '@angular/core';
import { Characters } from '../common/events/event-types/gen/cards/v1/cards_pb';

const CHARACTER_IMAGES: Partial<Record<Characters, string>> = {
  [Characters.SCARLET]: '/assets/characters/scarlet.jpg',
  [Characters.MUSTARD]: '/assets/characters/mustard.jpg',
  [Characters.WHITE]: '/assets/characters/white.jpg',
  [Characters.GREEN]: '/assets/characters/green.jpg',
  [Characters.PEACOCK]: '/assets/characters/peacock.jpg',
  [Characters.PLUM]: '/assets/characters/plum.jpg',
};

@Injectable({ providedIn: 'root' })
export class CardAccessService {
  getCharacterImage(character: Characters): string | undefined {
    return CHARACTER_IMAGES[character];
  }

  getCharacterName(character: Characters): string {
    return Characters[character];
  }
}
