import { Component, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { filter, map } from 'rxjs';
import { GoBackendServiceAdapter } from '../services/GoBackendServiceAdapter';
import { JoinSessionService } from '../services/JoinSessionService';

@Component({
  imports: [RouterOutlet, FormsModule],
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

  protected onJoinClick(): void {
    const name = this.username().trim();
    if (!name) {
      return;
    }

    this.joinError.set(null);

    this.joinSessionService.join(name).subscribe({
      next: (response) => {
        this.golangAdapter.joinSession(response.sessionId, response.username);
        void this.router.navigateByUrl('/test-message-bus');
      },
      error: (err) => {
        console.error('Failed to join session:', err);
        this.joinError.set('Could not join session. Please try again.');
      },
    });
  }
}
