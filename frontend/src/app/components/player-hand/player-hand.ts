import { ChangeDetectionStrategy, Component, signal } from '@angular/core';
import { ButtonDirective } from 'primeng/button';
import { Drawer } from 'primeng/drawer';
import { PlusIcon } from 'primeng/icons';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ButtonDirective, Drawer, PlusIcon],
  selector: 'app-player-hand',
  styleUrl: './player-hand.css',
  templateUrl: './player-hand.html',
})
export class PlayerHand {
  protected readonly visible = signal(false);
}
