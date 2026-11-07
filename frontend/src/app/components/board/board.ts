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
import { EndTurnSchema } from '../../common/events/event-types/gen/session/v1/turn_indication_pb';
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
  private readonly roomList = this.getEnumList(Rooms);
  private readonly suspectList = this.getEnumList(Characters);
  private readonly weaponList = this.getEnumList(Weapons);
  private readonly hallList = this.getEnumList(Halls);

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

    if (!isRoom || !key) {
      this.terminalService.sendResponse('Invalid move command format. Expected "move <room|hall> <key>".');
      return;
    }

    if (roomLower === 'hall') {
      const hall = Halls[keyUpper as keyof typeof Halls];
      if (typeof hall !== 'number' || hall === Halls.UNSPECIFIED) {
        this.terminalService.sendResponse(`Invalid hall ${keyUpper}. Use listhalls for valid halls`);
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
        this.terminalService.sendResponse(`Invalid room ${keyUpper}. Use listrooms for valid rooms`);
        return;
      }

      this.eventBus.publish({
        type: MessageType.Move,
        producer: EventProducer.CLIENT,
        payload: create(MoveSchema, { moveTo: { case: 'room', value: room } }),
      });

      return;
    }

    this.terminalService.sendResponse(`Invalid room type ${roomLower}. Must be either "hall" or "room".`);
    return;

  }

  private sendSuggest(command: string): void {

    if (!this.verifySuggestAccuseCommand(command)) {
      return;
    }

    const [_, character, weapon, room] = command.split(' ');

    const charVal = Characters[character.toUpperCase() as keyof typeof Characters];
    const roomVal = Rooms[room.toUpperCase() as keyof typeof Rooms];
    const weaponVal = Weapons[weapon.toUpperCase() as keyof typeof Weapons];


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

    if (!this.verifySuggestAccuseCommand(command)) {
      return;
    }

    const [_, character, weapon, room] = command.split(' ');

    const charVal = Characters[character.toUpperCase() as keyof typeof Characters];
    const roomVal = Rooms[room.toUpperCase() as keyof typeof Rooms];
    const weaponVal = Weapons[weapon.toUpperCase() as keyof typeof Weapons];

    this.eventBus.publish({
      type: MessageType.AccusationRequest,
      producer: EventProducer.CLIENT,
      payload: create(AccusationRequestSchema, {
        suggestion: { character: charVal, room: roomVal, weapon: weaponVal },
      }),
    });
  }

  private verifySuggestAccuseCommand(command: string): boolean {
    const [action, character, weapon, room] = command.split(' ');

    if (!character || !weapon || !room) {
      this.terminalService.sendResponse(`Missing arguments. Expected format ${action} <character> <weapon> <room>. Use listsuspects, listweapons, and listrooms for valid values`);
      return false;
    }

    const charVal = Characters[character.toUpperCase() as keyof typeof Characters];
    const roomVal = Rooms[room.toUpperCase() as keyof typeof Rooms];
    const weaponVal = Weapons[weapon.toUpperCase() as keyof typeof Weapons];

    if (typeof charVal !== 'number' || charVal === Characters.UNSPECIFIED) {
      this.terminalService.sendResponse(`Invalid character ${character}. Use listsuspects for valid characters`);
      return false;
    }

    if (typeof roomVal !== 'number' || roomVal === Rooms.UNSPECIFIED) {
      this.terminalService.sendResponse(`Invalid room ${room}. Use listrooms for valid rooms`);
      return false;
    }

    if (typeof weaponVal !== 'number' || weaponVal === Weapons.UNSPECIFIED) {
      this.terminalService.sendResponse(`Invalid weapon ${weapon}. Use listweapons for valid weapons`);
      return false;
    }

    return true;
  }

  private sendEndTurn(): void {

    this.eventBus.publish({
      type: MessageType.EndTurn,
      producer: EventProducer.CLIENT,
      payload: create(EndTurnSchema, {}),
    });
  }

  private getEnumList(enumObject: Record<number, string>): string[] {
    return Object.keys(enumObject)
    .filter((key) => key !== 'UNSPECIFIED' && Number.isNaN(Number(key)));
  }

}