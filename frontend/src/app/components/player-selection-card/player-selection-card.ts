import { ChangeDetectionStrategy, Component, computed, inject, input } from '@angular/core';
import { create } from '@bufbuild/protobuf';
import { ConfirmationService } from 'primeng/api';
import { ConfirmDialog } from 'primeng/confirmdialog';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import { Characters } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { CharacterSelectRequestSchema } from '../../common/events/event-types/gen/glm/v1/glm_pb';

// Text color is picked per card so the label stays readable on light and dark fills.
const CHARACTER_COLORS: Record<
  Exclude<Characters, Characters.UNSPECIFIED>,
  { background: string; text: string }
> = {
  [Characters.SCARLET]: { background: '#c1121f', text: '#fff' },
  [Characters.MUSTARD]: { background: '#e0a800', text: '#000' },
  [Characters.WHITE]: { background: '#f5f5f5', text: '#000' },
  [Characters.GREEN]: { background: '#2e7d32', text: '#fff' },
  [Characters.PEACOCK]: { background: '#1e4fa3', text: '#fff' },
  [Characters.PLUM]: { background: '#6a3d7a', text: '#fff' },
};

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ConfirmDialog],
  providers: [ConfirmationService],
  selector: 'app-player-selection-card',
  styleUrl: './player-selection-card.css',
  templateUrl: './player-selection-card.html',
})
export class PlayerSelectionCard {
  readonly character = input.required<Exclude<Characters, Characters.UNSPECIFIED>>();

  protected readonly color = computed(() => CHARACTER_COLORS[this.character()]);
  protected readonly name = computed(() => Characters[this.character()]);

  private readonly confirmationService = inject(ConfirmationService);
  private readonly eventBus = inject(EventBus);

  protected confirmSelection(): void {
    this.confirmationService.confirm({
      header: `Choose ${this.name()}?`,
      message: ' ',
      closable: false,
      rejectLabel: 'Cancel',
      rejectButtonProps: { label: 'Cancel', severity: 'secondary', outlined: true },
      acceptLabel: 'Select',
      acceptButtonProps: { label: 'Select', severity: 'success' },
      accept: () => {
        this.eventBus.publish({
          type: MessageType.CharacterSelectRequest,
          producer: EventProducer.CLIENT,
          payload: create(CharacterSelectRequestSchema, { character: this.character() }),
        });
      },
    });
  }
}
