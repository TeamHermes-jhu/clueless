import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TextMessageStore } from '../../common/data-stores/TextMessageStore';
import { PostmanTestServiceAdapter } from '../../services/PostmanTestServiceAdapter';

@Component({
  imports: [RouterLink],
  selector: 'app-test-message-bus',
  templateUrl: './test-message-bus.html',
})
export class TestMessageBus {
  // Injected before the adapters so it is listening before any connection opens.
  private readonly store = inject(TextMessageStore);
  // Injected so the websocket connections are opened when this page loads.
  private readonly postmanAdapter = inject(PostmanTestServiceAdapter);

  protected readonly messages = this.store.getAll();
}
