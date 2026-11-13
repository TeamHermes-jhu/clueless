import { ChangeDetectionStrategy, Component, computed, inject, input } from '@angular/core';
import { create } from '@bufbuild/protobuf';
import { ConfirmationService } from 'primeng/api';
import { ConfirmDialog } from 'primeng/confirmdialog';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import type { Characters } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { CharacterSelectRequestSchema } from '../../common/events/event-types/gen/glm/v1/glm_pb';
import { CardAccessService } from '../../services/CardAccessService';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ConfirmDialog],
  providers: [ConfirmationService],
  selector: 'app-player-selection-card',
  styleUrl: './player-selection-card.css',
  templateUrl: './player-selection-card.html',
})
export class PlayerSelectionCard {
  readonly character = input.required<Characters>();
  readonly selectable = input(true);

  private readonly cardAccessService = inject(CardAccessService);

  protected readonly image = computed(() =>
    this.cardAccessService.getCharacterImage(this.character()),
  );

  protected readonly name = computed(() =>
    this.cardAccessService.getCharacterName(this.character()),
  );

  private readonly confirmationService = inject(ConfirmationService);
  private readonly eventBus = inject(EventBus);

  protected confirmSelection(): void {
    if (!this.selectable()) {
      return;
    }
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
