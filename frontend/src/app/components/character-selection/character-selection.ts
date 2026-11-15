import { ChangeDetectionStrategy, ChangeDetectorRef, Component, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { filter } from 'rxjs';
import { PlayerStore } from '../../common/data-stores/PlayerStore';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import { Characters } from '../../common/events/event-types/gen/cards/v1/cards_pb';
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
  private readonly playerStore = inject(PlayerStore);
  protected readonly selectedCharacter = this.playerStore.character;

  private pendingCharacter: Characters | null = null;

  constructor() {
    const eventBus = inject(EventBus);
    const router = inject(Router);
    const changeDetector = inject(ChangeDetectorRef);

    eventBus
      .subscribe(MessageType.CharacterSelectRequest)
      .pipe(
        filter(({ event }) => event.producer === EventProducer.CLIENT),
        takeUntilDestroyed(),
      )
      .subscribe(({ event }) => {
        this.pendingCharacter = event.payload.character;
      });

    eventBus
      .on(MessageType.CharacterSelectChoices)
      .pipe(takeUntilDestroyed())
      .subscribe(({ assignments }) => {
        const myId = this.playerStore.playerId();
        const takenByOthers = new Set(
          assignments.filter((a) => a.playerId !== myId).map((a) => a.character),
        );
        this.availableCharacters = new Set(
          [...this.characters].filter((c) => !takenByOthers.has(c)),
        );
        const mine = assignments.find((a) => a.playerId === myId);
        this.playerStore.setCharacter(mine?.character ?? null);
        changeDetector.markForCheck();

        if (this.pendingCharacter !== null && mine?.character === this.pendingCharacter) {
          this.pendingCharacter = null;
          this.messageService.add({
            severity: 'success',
            summary: 'Character selected',
            detail: 'Your character choice was accepted.',
          });
        }
      });

    eventBus
      .subscribe(MessageType.StartGame)
      .pipe(takeUntilDestroyed())
      .subscribe(() => {
        void router.navigateByUrl('/integrated-board');
      });

    eventBus
      .subscribe(MessageType.Message)
      .pipe(
        filter(({ event }) => event.producer === EventProducer.SERVER),
        takeUntilDestroyed(),
      )
      .subscribe(({ event }) => {
        this.messageService.add({
          severity: 'info',
          summary: 'Message from server',
          detail: event.payload.msg,
        });
      });
  }
}
