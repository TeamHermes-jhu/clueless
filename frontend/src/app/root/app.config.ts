import { provideHttpClient } from '@angular/common/http';
import type { ApplicationConfig } from '@angular/core';
import { inject, provideAppInitializer, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter, withRouterConfig } from '@angular/router';
import Aura from '@primeuix/themes/aura';
import { providePrimeNG } from 'primeng/config';
import { BoardStateStore } from '../common/data-stores/BoardStateStore';
import { CharacterSelectionStore } from '../common/data-stores/CharacterSelectionStore';
import { PlayerStore } from '../common/data-stores/PlayerStore';
import { EventBus } from '../common/events/event-bus/event-bus';
import { routes } from './app.routes';

// Injected from .env at build time by tools/ng-env.ts; undefined under a plain `ng` run.
declare const PRIMENG_LICENSE: string | undefined;

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    // 'computed' moves the browser history back to where it was when a guard cancels
    // a back/forward navigation, so the URL bar stays in sync with the screen.
    provideRouter(routes, withRouterConfig({ canceledNavigationResolution: 'computed' })),
    provideHttpClient(),
    providePrimeNG({
      theme: { preset: Aura },
      license: typeof PRIMENG_LICENSE === 'undefined' ? undefined : PRIMENG_LICENSE,
    }),
    { provide: EventBus, useFactory: () => new EventBus('app') },
    PlayerStore,
    provideAppInitializer(() => {
      inject(CharacterSelectionStore);
      inject(BoardStateStore);
    }),
  ],
};
