declare const BACKEND_URL: string | undefined;

const configured = typeof BACKEND_URL === 'undefined' ? '' : BACKEND_URL.trim();

/** Base URL for HTTP requests to the backend, without a trailing slash. */
export const BACKEND_HTTP_URL = (configured || 'http://localhost:8000').replace(/\/+$/, '');

/** Base URL for WebSocket connections (http -> ws, https -> wss). */
export const BACKEND_WS_URL = BACKEND_HTTP_URL.replace(/^http/, 'ws');
