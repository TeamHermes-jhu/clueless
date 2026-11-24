import type { Routes } from '@angular/router';
import { CharacterSelection } from '../components/character-selection/character-selection';
import { IntegratedBoard } from '../components/integrated-board/integrated-board';
import { LandingPage } from '../components/landing-page/landing-page';
import { TestMessageBus } from '../components/test-message-bus/test-message-bus';
import { blockHistoryNavigation } from './block-history-navigation-guard';

export const routes: Routes = [
  { path: '', component: LandingPage },
  { path: 'character', component: CharacterSelection, canDeactivate: [blockHistoryNavigation] },
  {
    path: 'integrated-board',
    component: IntegratedBoard,
    canDeactivate: [blockHistoryNavigation],
  },
  { path: 'test-message-bus', component: TestMessageBus },
];
