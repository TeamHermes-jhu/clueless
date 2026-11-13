import { Injectable } from '@angular/core';
import { Characters } from '../common/events/event-types/gen/cards/v1/cards_pb';

const CHARACTER_IMAGES: Partial<Record<Characters, string>> = {
  [Characters.SCARLET]: '/images/characters/scarlet.jpg',
  [Characters.MUSTARD]: '/images/characters/mustard.jpg',
  [Characters.WHITE]: '/images/characters/white.jpg',
  [Characters.GREEN]: '/images/characters/green.jpg',
  [Characters.PEACOCK]: '/images/characters/peacock.jpg',
  [Characters.PLUM]: '/images/characters/plum.jpg',
};

@Injectable({ providedIn: 'root' })
export class CardAccessService {
  private selectableCharacters = new Set<Characters>([
    Characters.SCARLET,
    Characters.MUSTARD,
    Characters.WHITE,
    Characters.GREEN,
    Characters.PEACOCK,
    Characters.PLUM,
  ]);

  getSelectableCharacters(): ReadonlySet<Characters> {
    return this.selectableCharacters;
  }

  setSelectableCharacters(characters: Iterable<Characters>): void {
    this.selectableCharacters = new Set(characters);
  }

  getCharacterImage(character: Characters): string | undefined {
    return CHARACTER_IMAGES[character];
  }

  getCharacterName(character: Characters): string {
    return Characters[character];
  }
}
