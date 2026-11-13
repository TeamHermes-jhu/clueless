import { ChangeDetectionStrategy, ChangeDetectorRef, Component, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { filter } from 'rxjs';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import { Characters } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { GoBackendServiceAdapter } from '../../services/GoBackendServiceAdapter';
import { PlayerSelectionCard } from '../player-selection-card/player-selection-card';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [PlayerSelectionCard],
  selector: 'app-character-selection',
  styleUrl: './character-selection.css',
  templateUrl: './character-selection.html',
})
export class CharacterSelection {
  protected readonly characters = new Set<Characters>([
    Characters.SCARLET,
    Characters.MUSTARD,
    Characters.WHITE,
    Characters.GREEN,
    Characters.PEACOCK,
    Characters.PLUM,
  ]);

  protected availableCharacters = new Set<Characters>(this.characters);

  private readonly messageService = inject(MessageService);
  private readonly playerId = inject(GoBackendServiceAdapter).playerId;

  constructor() {
    const eventBus = inject(EventBus);
    const router = inject(Router);
    const changeDetector = inject(ChangeDetectorRef);

    eventBus
      .on(MessageType.CharacterChoices)
      .pipe(takeUntilDestroyed())
      .subscribe(({ characters }) => {
        this.availableCharacters = new Set(characters);
        changeDetector.markForCheck();
      });

    eventBus
      .subscribe(MessageType.CharacterSelectRequest)
      .pipe(
        filter(({ event }) => event.producer === EventProducer.CLIENT),
        takeUntilDestroyed(),
      )
      .subscribe(() => {
        void router.navigateByUrl('/integrated-board');
      });

    eventBus
      .on(MessageType.CharacterSelectResponse)
      .pipe(takeUntilDestroyed())
      .subscribe(({ playerId, ok }) => {
        if (playerId && playerId !== this.playerId()) {
          return;
        }
        this.messageService.add(
          ok
            ? {
                severity: 'success',
                summary: 'Character selected',
                detail: 'Your character choice was accepted.',
              }
            : {
                severity: 'error',
                summary: 'Selection failed',
                detail: 'That character could not be selected.',
              },
        );
        // TODO: Advance page only if response is ok
      });
  }
}
