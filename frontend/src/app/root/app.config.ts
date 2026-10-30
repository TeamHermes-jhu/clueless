import type { ApplicationConfig } from '@angular/core';
import { provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import Aura from '@primeuix/themes/aura';
import { providePrimeNG } from 'primeng/config';
import { routes } from './app.routes';
import { EventBus } from '../common/events/event-bus/event-bus';

// Injected from .env at build time by tools/ng-env.ts; undefined under a plain `ng` run.
declare const PRIMENG_LICENSE: string | undefined;

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideHttpClient(),
    providePrimeNG({
      theme: { preset: Aura },
      license: typeof PRIMENG_LICENSE === 'undefined' ? undefined : PRIMENG_LICENSE,
    }),
    { provide: EventBus, useFactory: () => new EventBus('app') },
  ],
};
