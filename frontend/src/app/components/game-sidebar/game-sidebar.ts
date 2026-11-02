import { ChangeDetectionStrategy, Component } from '@angular/core';
import { ButtonDirective } from 'primeng/button';
import { BarsIcon } from 'primeng/icons';
import {
  Sidebar,
  SidebarAside,
  SidebarContent,
  SidebarHeader,
  SidebarPanel,
  SidebarSpacer,
  SidebarTrigger,
} from 'primeng/sidebar';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    BarsIcon,
    ButtonDirective,
    Sidebar,
    SidebarAside,
    SidebarContent,
    SidebarHeader,
    SidebarPanel,
    SidebarSpacer,
    SidebarTrigger,
  ],
  selector: 'app-game-sidebar',
  styleUrl: './game-sidebar.css',
  templateUrl: './game-sidebar.html',
})
export class GameSidebar {}
