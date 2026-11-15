import { inject } from '@angular/core';
import type { CanDeactivateFn } from '@angular/router';
import { Router } from '@angular/router';

// Rejects navigations started by the browser's back/forward buttons so players cannot
// leave the current game screen that way. Navigations made by the app itself still work.
export const blockHistoryNavigation: CanDeactivateFn<unknown> = () =>
  inject(Router).currentNavigation()?.trigger !== 'popstate';
