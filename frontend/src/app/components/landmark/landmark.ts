import { ChangeDetectionStrategy, Component } from '@angular/core';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  selector: 'app-landmark',
  styleUrl: './landmark.css',
  templateUrl: './landmark.html',
})
export class Landmark {}
