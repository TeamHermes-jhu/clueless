import { ChangeDetectionStrategy, Component, inject, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { create } from '@bufbuild/protobuf';
import { MessageService } from 'primeng/api';
import { Terminal, TerminalService } from 'primeng/terminal';
import { Toast } from 'primeng/toast';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { MessageType } from '../../common/events/event-types/AppEvents';
import { Characters, Rooms, Weapons } from '../../common/events/event-types/gen/cards/v1/enums_pb';
import { MoveSchema } from '../../common/events/event-types/gen/session/v1/session_pb';
import {
  AccusationRequestSchema,
  SuggestionRequestSchema,
} from '../../common/events/event-types/gen/suggestion/v1/suggestion_pb';
import { GoBackendServiceAdapter } from '../../services/GoBackendServiceAdapter';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Terminal, Toast],
  providers: [MessageService, TerminalService],
  selector: 'app-board',
  styleUrl: './board.css',
  templateUrl: './board.html',
})
export class Board {
  private readonly golangAdapter = inject(GoBackendServiceAdapter);
  private readonly messageService = inject(MessageService);
  private readonly terminalService = inject(TerminalService);
  private readonly terminal = viewChild.required(Terminal);
  private readonly availableCommands = ['help', 'about', 'clear', 'move', 'suggest', 'accuse'];

  constructor() {
    inject(EventBus)
      .on(MessageType.Error)
      .pipe(takeUntilDestroyed())
      .subscribe(({ error }) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: error });
      });

    this.terminalService.commandHandler.pipe(takeUntilDestroyed()).subscribe((command) => {
      switch (command.trim()) {
        case 'clear':
          this.terminal().commands = [];
          break;
        case 'about':
          this.terminalService.sendResponse(`Welcome to Clue-less!`);
          break;
        case 'help':
          this.terminalService.sendResponse(`Available commands: ${this.availableCommands.join(', ')}`);
          break;
        case 'move':
          this.sendMove();
          break;
        case 'suggest':
          this.sendSuggest();
          break;
        case 'accuse':
          this.sendAccuse();
          break;
        default:
          this.terminalService.sendResponse(`Unknown command: ${command}`);
      }
    });
  }

  private sendMove(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Move before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'move',
      value: create(MoveSchema, {
        playerId,
        moveTo: { case: 'room', value: Rooms.STUDY },
      }),
    });
  }

  private sendSuggest(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Suggest before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'suggest',
      value: create(SuggestionRequestSchema, {
        playerId,
        character: Characters.WHITE,
        room: Rooms.HALL,
        weapon: Weapons.ROPE,
      }),
    });
  }

  private sendAccuse(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Suggest before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'accuse',
      value: create(AccusationRequestSchema, {
        suggestion: {
          playerId,
          character: Characters.WHITE,
          room: Rooms.HALL,
          weapon: Weapons.ROPE,
        },
      }),
    });
  }
}
