import type { Routes } from '@angular/router';
import { CharacterSelection } from '../components/character-selection/character-selection';
import { IntegratedBoard } from '../components/integrated-board/integrated-board';
import { TestMessageBus } from '../components/test-message-bus/test-message-bus';

export const routes: Routes = [
  { path: 'character', component: CharacterSelection },
  { path: 'integrated-board', component: IntegratedBoard },
  { path: 'test-message-bus', component: TestMessageBus },
];
