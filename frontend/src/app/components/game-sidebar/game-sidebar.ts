import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { create } from '@bufbuild/protobuf';
import { ButtonDirective } from 'primeng/button';
import { BarsIcon } from 'primeng/icons';
import {
  Sidebar,
  SidebarAside,
  SidebarContent,
  SidebarHeader,
  SidebarPanel,
  SidebarSpacer,
  SidebarTrigger,
} from 'primeng/sidebar';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { EventProducer, MessageType } from '../../common/events/event-types/AppEvents';
import { GetSessionInfoRequestSchema } from '../../common/events/event-types/gen/session/v1/session_pb';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    BarsIcon,
    ButtonDirective,
    CommonModule,
    Sidebar,
    SidebarAside,
    SidebarContent,
    SidebarHeader,
    SidebarPanel,
    SidebarSpacer,
    SidebarTrigger,
  ],
  selector: 'app-game-sidebar',
  styleUrl: './game-sidebar.css',
  templateUrl: './game-sidebar.html',
})

export class GameSidebar {

  private readonly eventBus = inject(EventBus);

  private suspects: string[]
  private weapons: string[]
  private rooms: string[]
  private players: string[]
  private stateSymbols: Record<number, string>
  private suspectGrid: number[][]
  private weaponGrid: number[][]
  private roomGrid: number[][]

  constructor() {

    this.suspects = [
      'Miss Scarlett',
      'Colonel Mustard',
      'Mrs. White',
      'Mr. Green',
      'Mrs. Peacock',
      'Professor Plum'
    ]
    this.weapons = [
      'Candlestick',
      'Dagger',
      'Lead Pipe',
      'Revolver',
      'Rope',
      'Wrench'
    ]
    this.rooms = [
      'Kitchen',
      'Ballroom',
      'Conservatory',
      'Dining Room',
      'Billiard Room',
      'Library',
      'Lounge',
      'Hall',
      'Study'
    ]

    this.stateSymbols = {
      0: '',
      1: '✗',
      2: '✓'
    };

    this.suspectGrid = [];
    this.weaponGrid = [];
    this.roomGrid = [];
    this.players = [];

    this.eventBus
    .on(MessageType.Session)
    .pipe(takeUntilDestroyed())
    .subscribe((session) => {
        this.createNotebooks(session);
    });
    this.getPlayerIds();

  }

  // Click handler
  toggleCell(rowIndex: number, colIndex: number, gridType?: 'suspect' | 'weapon' | 'room'): void {
    if (gridType === 'suspect') {
      this.suspectGrid[rowIndex][colIndex] = (this.suspectGrid[rowIndex][colIndex] + 1) % 3;
    } else if (gridType === 'weapon') {
      this.weaponGrid[rowIndex][colIndex] = (this.weaponGrid[rowIndex][colIndex] + 1) % 3;
    } else if (gridType === 'room') {
      this.roomGrid[rowIndex][colIndex] = (this.roomGrid[rowIndex][colIndex] + 1) % 3;
    }
  }

  private getPlayerIds(): void {
    this.eventBus.publish({
      type: MessageType.GetSessionInfoRequest,
      producer: EventProducer.CLIENT,
      payload: create(GetSessionInfoRequestSchema, {}),
    });
  }

  private createNotebooks(session: any): void {
    // Create notebooks for each playerthis.players = session.players;
      this.players = session.players;
      this.suspectGrid = this.suspects.map(() =>
        this.players.map(() => 0)
      );
      this.weaponGrid = this.weapons.map(() =>
        this.players.map(() => 0)
      );
      this.roomGrid = this.rooms.map(() =>
        this.players.map(() => 0)
      );
  }

}

