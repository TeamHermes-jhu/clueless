import { ChangeDetectionStrategy, ChangeDetectorRef, Component, inject, NgZone, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { create } from '@bufbuild/protobuf';
import { MessageService } from 'primeng/api';
import { Terminal, TerminalService } from 'primeng/terminal';
import { Toast } from 'primeng/toast';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { MessageType } from '../../common/events/event-types/AppEvents';
import { Characters, Rooms, Weapons, Halls } from '../../common/events/event-types/gen/cards/v1/cards_pb';
import { MoveSchema, ViewHandRequestSchema, ViewLocationsRequestSchema } from '../../common/events/event-types/gen/session/v1/session_pb';
import { EndTurnSchema } from '../../common/events/event-types/gen/session/v1/turn_indication_pb'
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
  private readonly zone = inject(NgZone);
  private readonly terminalService = inject(TerminalService);
  private readonly terminal = viewChild.required(Terminal);
  private readonly terminalChangeDetector = viewChild.required(Terminal, { read: ChangeDetectorRef });
  private readonly availableCommands = [
    'help', 
    'about', 
    'clear', 
    'move', 
    'suggest', 
    'accuse', 
    'endturn', 
    'viewhand', 
    'viewlocations', 
    'listrooms', 
    'listsuspects', 
    'listweapons'];
  private readonly roomList = Object.keys(Rooms)
    .filter((room) => room !== 'UNSPECIFIED' && Number.isNaN(Number(room)));
  private readonly suspectList = Object.keys(Characters)
    .filter((suspect) => suspect !== 'UNSPECIFIED' && Number.isNaN(Number(suspect)));
  private readonly weaponList = Object.keys(Weapons)
    .filter((weapon) => weapon !== 'UNSPECIFIED' && Number.isNaN(Number(weapon)));

  constructor() {
    inject(EventBus)
      .on(MessageType.Error)
      .pipe(takeUntilDestroyed())
      .subscribe(({ error }) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: error });
      });

    inject(EventBus)
      .on(MessageType.Message)
      .pipe(takeUntilDestroyed())
      .subscribe(({ msg }) => {
        this.zone.run(() => {
          const terminal = this.terminal();
          terminal.commands = [...terminal.commands, { text: 'server' }];
          this.terminalService.sendResponse(msg);
          this.terminalChangeDetector().detectChanges();
        });
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
        case 'endturn':
          this.sendEndTurn();
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
        case 'move':
          this.sendMove(command);
          break;
        case 'suggest':
          this.sendSuggest(command);
          break;
        case 'accuse':
          this.sendAccuse(command);
          break;
        default:
          this.terminalService.sendResponse(`Unknown command: ${commandName}`);
      }
    });
  }

  private sendViewHand(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send View Hand before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'viewHand',
      value: create(ViewHandRequestSchema, {
        playerId,
      }),
    });
  }

  private sendViewLocations(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send View Locations before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'viewLocations',
      value: create(ViewLocationsRequestSchema, {
        playerId,
      }),
    }); 

  }

  private sendMove(command: string): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Move before a session is joined.');
      return;
    }

    const [_, isRoom, key] = command.split(' ');
    const roomLower = isRoom.toLowerCase();
    const keyUpper = key.toUpperCase();

    if (roomLower === 'hall') {
      const hall = Halls[keyUpper as keyof typeof Halls];
      if (typeof hall !== 'number' || hall === Halls.UNSPECIFIED) {
        console.error('Invalid hall %s.', keyUpper);
        return;
      }

      this.golangAdapter.sendEnvelope({
        case: 'move',
        value: create(MoveSchema, {
          playerId,
          moveTo: { case: 'hall', value: hall },
        }),
      });

      return;
    }

    if (roomLower === 'room') {
      const room = Rooms[keyUpper as keyof typeof Rooms];
      if (typeof room !== 'number' || room === Rooms.UNSPECIFIED) {
        console.error('Invalid room %s.', keyUpper);
        return;
      }

      this.golangAdapter.sendEnvelope({
        case: 'move',
        value: create(MoveSchema, {
          playerId,
          moveTo: { case: 'room', value: room },
        }),
      });

      return;
    }

    console.error('Invalid room type %s. Must be either "hall" or "room".', roomLower);
    return;

  }

  private sendSuggest(command: string): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Suggest before a session is joined.');
      return;
    }

    const [_, character, room, weapon] = command.split(' ');

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

    this.golangAdapter.sendEnvelope({
      case: 'suggest',
      value: create(SuggestionRequestSchema, {
        playerId,
        character: charVal,
        room: roomVal,
        weapon: weaponVal,
      }),
    });
  }

  private sendAccuse(command: string): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Accuse before a session is joined.');
      return;
    }

    const [_, character, room, weapon] = command.split(' ');

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

    this.golangAdapter.sendEnvelope({
      case: 'accuse',
      value: create(AccusationRequestSchema, {
        suggestion: {
          playerId,
          character: charVal,
          room: roomVal,
          weapon: weaponVal,
        },
      }),
    });
  }

  private sendEndTurn(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send End Turn before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'endTurn',
      value: create(EndTurnSchema, {
        playerId,
      }),
    });
  }

}