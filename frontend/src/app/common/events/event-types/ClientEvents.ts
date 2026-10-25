import type { EventProducer, MessageType } from './AppEvents';

export interface TextMessageRecieve {
  type: MessageType.TextMessageRecieve;
  producer: EventProducer.SERVER;
  payload: { message: string };
}

export interface TextMessageSend {
  type: MessageType.TextMessageSend;
  producer: EventProducer.CLIENT;
  payload: { message: string };
}
