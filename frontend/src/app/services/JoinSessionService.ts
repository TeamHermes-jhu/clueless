import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { create, fromJsonString, toJsonString } from '@bufbuild/protobuf';
import { map } from 'rxjs';
import type { Observable } from 'rxjs';
import {
  JoinRequestSchema,
  JoinResponseSchema,
} from '../common/events/event-types/gen/session/v1/session_pb';
import type { JoinResponse } from '../common/events/event-types/gen/session/v1/session_pb';

@Injectable({ providedIn: 'root' })
export class JoinSessionService {
  private readonly http = inject(HttpClient);

  join(username: string): Observable<JoinResponse> {
    const request = create(JoinRequestSchema, { username });
    const body = toJsonString(JoinRequestSchema, request);

    return this.http
      .post('http://localhost:8000/join_session', body, {
        headers: { 'Content-Type': 'application/json' },
        responseType: 'text', // Server writes a JSON string body, not an object.
      })
      .pipe(map((responseBody) => fromJsonString(JoinResponseSchema, responseBody)));
  }
}
