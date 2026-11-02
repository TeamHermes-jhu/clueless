import { ChangeDetectionStrategy, Component, inject, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Terminal, TerminalService } from 'primeng/terminal';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Terminal],
  providers: [TerminalService],
  selector: 'app-board',
  styleUrl: './board.css',
  templateUrl: './board.html',
})
export class Board {
  private readonly terminalService = inject(TerminalService);
  private readonly terminal = viewChild.required(Terminal);

  constructor() {
    this.terminalService.commandHandler.pipe(takeUntilDestroyed()).subscribe((command) => {
      switch (command.trim()) {
        case 'clear':
          this.terminal().commands = [];
          break;
        case 'about':
          this.terminalService.sendResponse(`Welcome to Clue-less!`);
          break;
        default:
          this.terminalService.sendResponse(`Unknown command: ${command}`);
      }
    });
  }
}
