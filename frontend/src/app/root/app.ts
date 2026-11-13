import { Component, inject, signal } from '@angular/core';
import { takeUntilDestroyed, toSignal } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { MessageService } from 'primeng/api';
import { Toast } from 'primeng/toast';
import { filter, map } from 'rxjs';
import { EventBus } from '../common/events/event-bus/event-bus';
import { MessageType } from '../common/events/event-types/AppEvents';
import { GoBackendServiceAdapter } from '../services/GoBackendServiceAdapter';
import { JoinSessionService } from '../services/JoinSessionService';

@Component({
  imports: [RouterOutlet, FormsModule, Toast],
  providers: [MessageService],
  selector: 'app-root',
  styleUrl: './app.css',
  templateUrl: './app.html',
})
export class App {
  protected readonly title = signal('clueless-frontend-app');

  // Hide the main page content while a routed page is shown.
  protected readonly onMainPage = toSignal(
    inject(Router).events.pipe(
      filter((e) => e instanceof NavigationEnd),
      map((e) => e.urlAfterRedirects === '/'),
    ),
    { initialValue: true },
  );

  protected readonly username = signal('');
  protected readonly joinError = signal<string | null>(null);

  private readonly joinSessionService = inject(JoinSessionService);
  private readonly golangAdapter = inject(GoBackendServiceAdapter);
  private readonly router = inject(Router);
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

  protected onJoinClick(): void {
    const name = this.username().trim();
    if (!name) {
      return;
    }

    this.joinError.set(null);

    this.joinSessionService.join(name).subscribe({
      next: (response) => {
        this.golangAdapter.joinSession(response.sessionId, response.username);
        void this.router.navigateByUrl('/character');
      },
      error: (err) => {
        console.error('Failed to join session:', err);
        this.joinError.set('Could not join session. Please try again.');
      },
    });
  }
}
