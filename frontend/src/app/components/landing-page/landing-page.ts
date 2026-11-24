import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ButtonDirective } from 'primeng/button';
import { InputText } from 'primeng/inputtext';
import { GoBackendServiceAdapter } from '../../services/GoBackendServiceAdapter';
import { JoinSessionService } from '../../services/JoinSessionService';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, InputText, ButtonDirective],
  selector: 'app-landing-page',
  styleUrl: './landing-page.css',
  templateUrl: './landing-page.html',
})
export class LandingPage {
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
        void this.router.navigateByUrl('/character');
      },
      error: (err) => {
        console.error('Failed to join session:', err);
        this.joinError.set('Could not join session. Please try again.');
      },
    });
  }
}
