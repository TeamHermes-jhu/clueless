import { ChangeDetectionStrategy, Component } from '@angular/core';
import { Board } from '../board/board';
import { GameTerminal } from '../game-terminal/game-terminal';
import { GameSidebar } from '../game-sidebar/game-sidebar';
import { Landmark } from '../landmark/landmark';
import { PlayerHand } from '../player-hand/player-hand';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Board, GameSidebar, GameTerminal, Landmark, PlayerHand],
  selector: 'app-integrated-board',
  styleUrl: './integrated-board.css',
  templateUrl: './integrated-board.html',
})
export class IntegratedBoard {}
