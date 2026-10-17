import { Component, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { map, scan } from 'rxjs';
import { EventBus } from '../../common/events/event-bus/event-bus';
import { MessageType } from '../../common/events/event-types/AppEvents';
import { JsonServiceAdapter } from '../../services/JsonServiceAdapter';

@Component({
  imports: [RouterLink],
  selector: 'app-test-message-bus',
  templateUrl: './test-message-bus.html',
})
export class TestMessageBus {
  private readonly eventBus = inject(EventBus);
  // Injected so the websocket connection is opened when this page loads.
  private readonly adapter = inject(JsonServiceAdapter);

  protected readonly messages = toSignal(
    this.eventBus.on(MessageType.TextMessageRecieve).pipe(
      map(payload => payload.message),
      scan((all, message) => [...all, message], [] as string[]),
    ),
    { initialValue: [] },
  );
}
