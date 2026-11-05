import { ChangeDetectionStrategy, Component, inject, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { create } from '@bufbuild/protobuf';
import { MessageService } from 'primeng/api';
import { Terminal, TerminalService } from 'primeng/terminal';
import { Toast } from 'primeng/toast';
import { filter } from 'rxjs';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import { Characters, Rooms, Weapons, Halls } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { MoveSchema, ViewHandRequestSchema, ViewLocationsRequestSchema } from '../../common/events/event-types/gen/session/v1/session_pb';
import { EndTurnSchema } from '../../common/events/event-types/gen/session/v1/turn_indication_pb'
import {
  AccusationRequestSchema,
  SuggestionRequestSchema,
} from '../../common/events/event-types/gen/suggestion/v1/suggestion_pb';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Terminal, Toast],
  providers: [MessageService, TerminalService],
  selector: 'app-board',
  styleUrl: './board.css',
  templateUrl: './board.html',
})
export class Board {
  private readonly eventBus = inject(EventBus);
  private readonly messageService = inject(MessageService);
  private readonly terminalService = inject(TerminalService);
  private readonly terminal = viewChild.required(Terminal);
  private readonly availableCommands = [
    'help', 
    'about', 
    'clear', 
    'selectcharacter',
    'move', 
    'suggest', 
    'accuse', 
    'respond',
    'endturn', 
    'viewhand', 
    'viewlocations', 
    'listrooms', 
    'listsuspects', 
    'listweapons',
    'listhalls'
  ];
  private readonly roomList = Object.keys(Rooms)
    .filter((room) => room !== 'UNSPECIFIED' && Number.isNaN(Number(room)));
  private readonly suspectList = Object.keys(Characters)
    .filter((suspect) => suspect !== 'UNSPECIFIED' && Number.isNaN(Number(suspect)));
  private readonly weaponList = Object.keys(Weapons)
    .filter((weapon) => weapon !== 'UNSPECIFIED' && Number.isNaN(Number(weapon)));
  private readonly hallList = Object.keys(Halls)
    .filter((hall) => hall !== 'UNSPECIFIED' && Number.isNaN(Number(hall)));

  constructor() {
    // Subscribe to error events to get Toast in top corner
    this.eventBus
      .on(MessageType.Error)
      .pipe(takeUntilDestroyed())
      .subscribe(({ error }) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: error });
      });

    // Subscribe to message events to get messages from server in terminal
    this.eventBus
      .subscribe(MessageType.Message)
      .pipe(
        filter(({ event }) => event.producer === EventProducer.SERVER),
        takeUntilDestroyed(),
      )
      .subscribe(({ event }) => {
        this.printToTerminal(event.payload.msg);
      });

    this.terminalService.commandHandler.pipe(takeUntilDestroyed()).subscribe((command) => {
      // execute the command based on the user's input
      const commandName = command.trim().split(' ')[0];
      switch (commandName) {
        case 'clear':
          this.terminal().commands = [];
          break;
        case 'about':
          this.terminalService.sendResponse(`Welcome to Clue-less!`);
          break;
        case 'help':
          this.terminalService.sendResponse(`Available commands: ${this.availableCommands.join(', ')}`);
          break;
        case 'viewhand':
          this.sendViewHand();
          break;
        case 'viewlocations':
          this.sendViewLocations();
          break;
        case 'listrooms':
          this.terminalService.sendResponse(`Available rooms: ${this.roomList.join(', ')}`);
          break;
        case 'listsuspects':
          this.terminalService.sendResponse(`Available suspects: ${this.suspectList.join(', ')}`);
          break;
        case 'listweapons':
          this.terminalService.sendResponse(`Available weapons: ${this.weaponList.join(', ')}`);
          break;
        case 'listhalls':
          this.terminalService.sendResponse(`Available halls: ${this.hallList.join(', ')}`);
          break;
        case 'move':
          this.sendMove(command);
          break;
        case 'suggest':
          this.sendSuggest(command);
          break;
        case 'accuse':
          this.sendAccuse(command);
          break;
        case 'endturn':
          this.sendEndTurn();
          break;
        case 'selectcharacter':
          this.terminalService.sendResponse(`TODO: Implement character selection`);
          break;
        case 'respond':
          this.terminalService.sendResponse(`TODO: Implement response handling`);
          break;
        default:
          this.terminalService.sendResponse(`Unknown command: ${commandName}`);
      }
    });
  }

    private printToTerminal(text: string): void {
      const terminal = this.terminal();
      terminal.commands.push({ text: '' });
      this.terminalService.sendResponse(text);
      terminal.cd.markForCheck();
    }

  private sendViewHand(): void {

    this.eventBus.publish({
      type: MessageType.ViewHandRequest,
      producer: EventProducer.CLIENT,
      payload: create(ViewHandRequestSchema, {}),
    });
  }

  private sendViewLocations(): void {

    this.eventBus.publish({
      type: MessageType.ViewLocationsRequest,
      producer: EventProducer.CLIENT,
      payload: create(ViewLocationsRequestSchema, {}),
    });

  }

  private sendMove(command: string): void {

    const [_, isRoom, key] = command.split(' ');
    const roomLower = isRoom.toLowerCase();
    const keyUpper = key.toUpperCase();

    if (roomLower === 'hall') {
      const hall = Halls[keyUpper as keyof typeof Halls];
      if (typeof hall !== 'number' || hall === Halls.UNSPECIFIED) {
        console.error('Invalid hall %s.', keyUpper);
        return;
      }

      this.eventBus.publish({
        type: MessageType.Move,
        producer: EventProducer.CLIENT,
        payload: create(MoveSchema, { moveTo: { case: 'hall', value: hall } }),
      });

      return;
    }

    if (roomLower === 'room') {
      const room = Rooms[keyUpper as keyof typeof Rooms];
      if (typeof room !== 'number' || room === Rooms.UNSPECIFIED) {
        console.error('Invalid room %s.', keyUpper);
        return;
      }

      this.eventBus.publish({
        type: MessageType.Move,
        producer: EventProducer.CLIENT,
        payload: create(MoveSchema, { moveTo: { case: 'room', value: room } }),
      });

      return;
    }

    console.error('Invalid room type %s. Must be either "hall" or "room".', roomLower);
    return;

  }

  private sendSuggest(command: string): void {

    const [_, character, weapon, room] = command.split(' ');

    const charVal = Characters[character.toUpperCase() as keyof typeof Characters];
    const roomVal = Rooms[room.toUpperCase() as keyof typeof Rooms];
    const weaponVal = Weapons[weapon.toUpperCase() as keyof typeof Weapons];

    if (typeof charVal !== 'number' || charVal === Characters.UNSPECIFIED) {
      console.error('Invalid character %s.', character);
      return;
    }

    if (typeof roomVal !== 'number' || roomVal === Rooms.UNSPECIFIED) {
      console.error('Invalid room %s.', room);
      return;
    }

    if (typeof weaponVal !== 'number' || weaponVal === Weapons.UNSPECIFIED) {
      console.error('Invalid weapon %s.', weapon);
      return;
    }


    this.eventBus.publish({
      type: MessageType.SuggestionRequest,
      producer: EventProducer.CLIENT,
      payload: create(SuggestionRequestSchema, {
        character: charVal,
        room: roomVal,
        weapon: weaponVal,
      }),
    });
  }

  private sendAccuse(command: string): void {

    const [_, character, weapon, room] = command.split(' ');

    const charVal = Characters[character.toUpperCase() as keyof typeof Characters];
    const roomVal = Rooms[room.toUpperCase() as keyof typeof Rooms];
    const weaponVal = Weapons[weapon.toUpperCase() as keyof typeof Weapons];

    if (typeof charVal !== 'number' || charVal === Characters.UNSPECIFIED) {
      console.error('Invalid character %s.', character);
      return;
    }

    if (typeof roomVal !== 'number' || roomVal === Rooms.UNSPECIFIED) {
      console.error('Invalid room %s.', room);
      return;
    }

    if (typeof weaponVal !== 'number' || weaponVal === Weapons.UNSPECIFIED) {
      console.error('Invalid weapon %s.', weapon);
      return;
    }

    this.eventBus.publish({
      type: MessageType.AccusationRequest,
      producer: EventProducer.CLIENT,
      payload: create(AccusationRequestSchema, {
        suggestion: { character: charVal, room: roomVal, weapon: weaponVal },
      }),
    });
  }

  private sendEndTurn(): void {

    this.eventBus.publish({
      type: MessageType.EndTurn,
      producer: EventProducer.CLIENT,
      payload: create(EndTurnSchema, {}),
    });
  }

}