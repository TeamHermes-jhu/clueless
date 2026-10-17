import { Injectable } from "@angular/core";
import type { AppEvent} from "../common/events/event-types/AppEvents";
import { EventProducer, MessageType } from "../common/events/event-types/AppEvents";
import { AbstractServiceAdapter } from "./AbstractServiceAdapter";

type JsonObject = Record<string, unknown>;

function isObject(value: unknown): value is JsonObject {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}

@Injectable({ providedIn: 'root' })
export class JsonServiceAdapter extends AbstractServiceAdapter {

    constructor() {
        super('localhost', 8080, '/ws');
    }

    protected override parse(raw: unknown): AppEvent | null {
        if (typeof raw !== 'string') {
            return null;
        }

        const data: unknown = JSON.parse(raw);
        if (!isObject(data) || !isObject(data['payload'])) {
            return null;
        }
        const payload = data['payload'];

        switch (data['type']) {
            case MessageType.TextMessageRecieve:
                if (typeof payload['message'] !== 'string') {
                    return null;
                }
                return {
                    type: MessageType.TextMessageRecieve,
                    producer: EventProducer.SERVER,
                    payload: { message: payload['message'] },
                };
            default:
                return null;
        }
    }
}
