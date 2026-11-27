import type { Signal, WritableSignal } from '@angular/core';
import { inject, Injectable, signal } from '@angular/core';
import { Characters } from '../common/events/event-types/gen/cards/v1/cards_pb';
import { CardAccessService } from './CardAccessService';

const ICON_SIZE = 128;
const RING_WIDTH = 10;
const CROP_FRACTION = 0.6;

const CHARACTER_COLORS: Partial<Record<Characters, string>> = {
  [Characters.SCARLET]: '#c8102e',
  [Characters.MUSTARD]: '#e1ad01',
  [Characters.WHITE]: '#f5f5f5',
  [Characters.GREEN]: '#1f7a3a',
  [Characters.PEACOCK]: '#1f4fbf',
  [Characters.PLUM]: '#6b2c91',
};

@Injectable({ providedIn: 'root' })
export class PlayerIconAccessService {
  private readonly cardAccessService = inject(CardAccessService);
  private readonly icons = new Map<Characters, WritableSignal<string | undefined>>();

  getIcon(character: Characters): Signal<string | undefined> {
    let icon = this.icons.get(character);
    if (!icon) {
      const created = signal<string | undefined>(undefined);
      icon = created;
      this.icons.set(character, created);

      const card = this.cardAccessService.getCharacterImage(character);
      if (card) {
        this.createIcon(card, this.getColor(character)).then(
          (url) => created.set(url),
          (err: unknown) => {
            console.warn(`Could not build icon for ${Characters[character]}; using the card.`, err);
            created.set(card);
          },
        );
      }
    }
    return icon.asReadonly();
  }

  getColor(character: Characters): string {
    return CHARACTER_COLORS[character] ?? '#808080';
  }

  getName(character: Characters): string {
    return this.cardAccessService.getCharacterName(character);
  }

  private async createIcon(cardUrl: string, ringColor: string): Promise<string> {
    const card = new Image();
    card.src = cardUrl;
    await card.decode();

    const canvas = document.createElement('canvas');
    canvas.width = ICON_SIZE;
    canvas.height = ICON_SIZE;
    const context = canvas.getContext('2d');
    if (!context) {
      throw new Error('2D canvas is not available');
    }

    const side = Math.min(card.naturalWidth, card.naturalHeight) * CROP_FRACTION;
    const sourceX = (card.naturalWidth - side) / 2;
    const sourceY = (card.naturalHeight - side) / 2;
    const center = ICON_SIZE / 2;

    context.save();
    context.beginPath();
    context.arc(center, center, center, 0, Math.PI * 2);
    context.clip();
    context.drawImage(card, sourceX, sourceY, side, side, 0, 0, ICON_SIZE, ICON_SIZE);
    context.restore();

    context.beginPath();
    context.arc(center, center, center - RING_WIDTH / 2, 0, Math.PI * 2);
    context.lineWidth = RING_WIDTH;
    context.strokeStyle = ringColor;
    context.stroke();

    return canvas.toDataURL('image/png');
  }
}
