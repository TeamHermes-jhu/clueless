import { ChangeDetectionStrategy, Component, inject, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { create } from '@bufbuild/protobuf';
import { Terminal, TerminalService } from 'primeng/terminal';
import { filter } from 'rxjs';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import type { Card } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { Characters, Rooms, Weapons, Halls } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { MoveSchema, 
  ViewHandRequestSchema, 
  ViewLocationsRequestSchema,
  ViewCharactersRequestSchema, 
} from '../../common/events/event-types/gen/session/v1/session_pb';
import { EndTurnSchema } from '../../common/events/event-types/gen/session/v1/turn_indication_pb';
import {
  AccusationRequestSchema,
  SuggestionRequestSchema,
  SuggestionResponseSchema,
} from '../../common/events/event-types/gen/suggestion/v1/suggestion_pb';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Terminal],
  providers: [TerminalService],
  selector: 'app-game-terminal',
  styleUrl: './game-terminal.css',
  templateUrl: './game-terminal.html',
})
export class GameTerminal {
  private readonly eventBus = inject(EventBus);
  private readonly terminalService = inject(TerminalService);
  private readonly terminal = viewChild.required(Terminal);
  private readonly availableCommands = [
    'help', 
    'about', 
    'clear', 
    'move', 
    'suggest', 
    'accuse', 
    'respond',
    'endturn', 
    'viewhand', 
    'viewlocations', 
    'viewcharacters',
    'listrooms', 
    'listsuspects', 
    'listweapons',
    'listhalls'
  ];
  private readonly roomList = this.getEnumList(Rooms);
  private readonly suspectList = this.getEnumList(Characters);
  private readonly weaponList = this.getEnumList(Weapons);
  private readonly hallList = this.getEnumList(Halls);
  private pendingDisprove: { suggestionId: number; validCards: Card[] } | null = null;

  constructor() {
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

    this.eventBus
      .on(MessageType.DisproveRequest)
      .pipe(takeUntilDestroyed())
      .subscribe((disproveRequest) => {
        this.pendingDisprove = {
          suggestionId: disproveRequest.suggestion?.suggestionId ?? 0,
          validCards: disproveRequest.validCards,
        };

        const options = disproveRequest.validCards.map((card) => this.cardName(card)).join(', ');
        this.printToTerminal(`You can disprove with: ${options}. Use "disprove <card>".`);
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
        case 'viewcharacters':
          this.sendViewCharacters();
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
        case 'disprove':
          this.sendDisprove(command);
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

  private sendViewCharacters(): void {

    this.eventBus.publish({
      type: MessageType.ViewCharactersRequest,
      producer: EventProducer.CLIENT,
      payload: create(ViewCharactersRequestSchema, {}),
    });

  }

  private sendMove(command: string): void {

    const [_, isRoom, key] = command.split(' ');

    if (!isRoom || !key) {
      this.terminalService.sendResponse('Invalid move command format. Expected "move <room|hall> <key>".');
      return;
    }

    const roomLower = isRoom.toLowerCase();
    const keyUpper = key.toUpperCase();

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

    if (typeof weaponVal !== 'number' || weaponVal === Weapons.UNSPECIFIED) {
      this.terminalService.sendResponse(`Invalid weapon ${weapon}. Use listweapons for valid weapons`);
      return false;
    }

    if (typeof roomVal !== 'number' || roomVal === Rooms.UNSPECIFIED) {
      this.terminalService.sendResponse(`Invalid room ${room}. Use listrooms for valid rooms`);
      return false;
    }

    return true;
  }

  private sendDisprove(command: string): void {

    if (!this.pendingDisprove) {
      this.terminalService.sendResponse('There is nothing to disprove right now.');
      return;
    }

    const [, cardArg] = command.split(' ');

    if (!cardArg) {
      this.terminalService.sendResponse(`Missing card. Expected format "disprove <card>".`);
      return;
    }

    const { suggestionId, validCards } = this.pendingDisprove;
    const card = validCards.find((candidate) => this.cardName(candidate).toLowerCase() === cardArg.toLowerCase());

    if (!card) {
      this.terminalService.sendResponse(
        `Invalid card ${cardArg}. Options: ${validCards.map((candidate) => this.cardName(candidate)).join(', ')}`,
      );
      return;
    }

    this.pendingDisprove = null;

    switch (card.card.case) {
      case 'character':
        this.publishSuggestionResponse(suggestionId, { case: 'character', value: card.card.value });
        return;
      case 'weapon':
        this.publishSuggestionResponse(suggestionId, { case: 'weapon', value: card.card.value });
        return;
      case 'room':
        this.publishSuggestionResponse(suggestionId, { case: 'room', value: card.card.value });
        return;
      default:
        return;
    }
  }

  private publishSuggestionResponse(
    suggestionId: number,
    card: { case: 'character'; value: Characters } | { case: 'weapon'; value: Weapons } | { case: 'room'; value: Rooms },
  ): void {
    this.eventBus.publish({
      type: MessageType.SuggestionResponse,
      producer: EventProducer.CLIENT,
      payload: create(SuggestionResponseSchema, {
        suggestionId,
        hasCard: true,
        card,
      }),
    });
  }

  private cardName(card: Card): string {
    switch (card.card.case) {
      case 'character':
        return Characters[card.card.value];
      case 'weapon':
        return Weapons[card.card.value];
      case 'room':
        return Rooms[card.card.value];
      default:
        return '';
    }
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
