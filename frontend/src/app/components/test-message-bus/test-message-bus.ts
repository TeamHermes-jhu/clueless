import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { create } from '@bufbuild/protobuf';
import { TextMessageStore } from '../../common/data-stores/TextMessageStore';
import { PostmanTestServiceAdapter } from '../../services/PostmanTestServiceAdapter';
import { GoBackendServiceAdapter } from '../../services/GoBackendServiceAdapter';
import { Rooms, Characters, Weapons } from '../../common/events/event-types/gen/cards/v1/enums_pb';
import { MoveSchema } from '../../common/events/event-types/gen/session/v1/session_pb';
import { SuggestionRequestSchema, AccusationRequestSchema } from '../../common/events/event-types/gen/suggestion/v1/suggestion_pb';

interface Action {
  label: string;
  send: () => void;
}

@Component({
  imports: [RouterLink],
  selector: 'app-test-message-bus',
  templateUrl: './test-message-bus.html',
})
export class TestMessageBus {
  // Injected before the adapters so it is listening before any connection opens.
  private readonly store = inject(TextMessageStore);
  // Injected so the websocket connections are opened when this page loads.
  // private readonly postmanAdapter = inject(PostmanTestServiceAdapter);
  private readonly golangAdapter = inject(GoBackendServiceAdapter)

  protected readonly messages = this.store.getAll();
  protected readonly connected = this.golangAdapter.connected;

  // TODO: these require a list of places/people/weapons to choose from
  protected readonly actions: readonly Action[] = [
    { label: 'Move', send: () => this.sendMove() },
    { label: 'Suggest', send: () => this.sendSuggest() },
    { label: 'Accuse', send: () => this.sendAccuse() },
  ];

  private sendMove(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Move before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'move',
      value: create(MoveSchema, {
        playerId,
        moveTo: { case: 'room', value: Rooms.STUDY },
      }),
    });
  }

  private sendSuggest(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Suggest before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'suggest',
      value: create(SuggestionRequestSchema, {
        playerId,
        character: Characters.WHITE,
        room: Rooms.HALL,
        weapon: Weapons.ROPE,
      }),
    });
  }

  private sendAccuse(): void {
    const playerId = this.golangAdapter.playerId();
    if (!playerId) {
      console.error('Cannot send Suggest before a session is joined.');
      return;
    }

    this.golangAdapter.sendEnvelope({
      case: 'accuse',
      value: create(AccusationRequestSchema, {
        suggestion: {
          playerId,
          character: Characters.WHITE,
          room: Rooms.HALL,
          weapon: Weapons.ROPE,
        },
      }),
    });
  }
}

