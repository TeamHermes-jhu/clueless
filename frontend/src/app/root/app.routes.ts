import type { Routes } from '@angular/router';
import { CharacterSelection } from '../components/character-selection/character-selection';
import { IntegratedBoard } from '../components/integrated-board/integrated-board';
import { TestMessageBus } from '../components/test-message-bus/test-message-bus';
import { blockHistoryNavigation } from './block-history-navigation-guard';

export const routes: Routes = [
  { path: 'character', component: CharacterSelection, canDeactivate: [blockHistoryNavigation] },
  {
    path: 'integrated-board',
    component: IntegratedBoard,
    canDeactivate: [blockHistoryNavigation],
  },
  { path: 'test-message-bus', component: TestMessageBus },
];
