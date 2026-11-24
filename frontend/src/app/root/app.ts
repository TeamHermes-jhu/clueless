import { Component, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterOutlet } from '@angular/router';
import { MessageService } from 'primeng/api';
import { Toast } from 'primeng/toast';
import { EventBus } from '../common/events/event-bus/event-bus';
import { MessageType } from '../common/events/event-types/AppEvents';
import { GoBackendServiceAdapter } from '../services/GoBackendServiceAdapter';

@Component({
  imports: [RouterOutlet, Toast],
  providers: [MessageService],
  selector: 'app-root',
  styleUrl: './app.css',
  templateUrl: './app.html',
})
export class App {
  protected readonly title = signal('clueless-frontend-app');

  private readonly golangAdapter = inject(GoBackendServiceAdapter);
  private readonly messageService = inject(MessageService);

  constructor() {
    // Show server errors as a toast on every page, not just the board.
    inject(EventBus)
      .on(MessageType.Error)
      .pipe(takeUntilDestroyed())
      .subscribe(({ error }) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: error });
      });
  }
}
