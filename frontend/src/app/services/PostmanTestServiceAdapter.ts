import { Injectable } from "@angular/core";
import type { AppEvent } from "../common/events/event-types/AppEvents";
import { EventProducer, MessageType } from "../common/events/event-types/AppEvents";
import { AbstractServiceAdapter } from "./AbstractServiceAdapter";

type JsonObject = Record<string, unknown>;

function isObject(value: unknown): value is JsonObject {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}

@Injectable({ providedIn: 'root' })
export class PostmanTestServiceAdapter extends AbstractServiceAdapter {

    constructor() {
        super('localhost', 8081, '/ws');
    }

    protected override parse(raw: unknown): AppEvent | null {
        if (typeof raw !== 'string') {
            return null;
        }

        let data: unknown;
        try {
            data = JSON.parse(raw);
        } catch {
            return this.textMessage(raw);
        }
        if (!isObject(data) || !isObject(data['payload'])) {
            console.warn('PostmanTestServiceAdapter ignored message without an object payload:', raw);
            return null;
        }
        const payload = data['payload'];

        // Allow the enum name ("TextMessageRecieve") or its numeric value (0) from Postman.
        const type = typeof data['type'] === 'string'
            ? MessageType[data['type'] as keyof typeof MessageType]
            : data['type'];

        switch (type) {
            case MessageType.TextMessageRecieve:
                return typeof payload['message'] === 'string' ? this.textMessage(payload['message']) : null;
            default:
                console.warn('PostmanTestServiceAdapter ignored unknown message type:', data['type']);
                return null;
        }
    }

    private textMessage(message: string): AppEvent {
        return {
            type: MessageType.TextMessageRecieve,
            producer: EventProducer.SERVER,
            payload: { message },
        };
    }
}
