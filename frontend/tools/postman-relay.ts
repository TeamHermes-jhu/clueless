// Local test relay so Postman can push messages to the Angular app.
//
// Postman and the browser are both WebSocket *clients*, so they need a server
// in the middle. Every message received from one client is forwarded to all
// other connected clients.
//
//   Run:      npm run relay   (Node runs this .ts file directly via type stripping)
//   Browser:  ws://localhost:8081/ws   (PostmanTestServiceAdapter)
//   Postman:  WebSocket request to ws://localhost:8081/ws
//             or HTTP POST http://localhost:8081/send with a JSON body
import { createServer } from 'node:http';
import type { WebSocket } from 'ws';
import { WebSocketServer } from 'ws';

const PORT = Number(process.env['RELAY_PORT'] ?? 8081);

const server = createServer((req, res) => {
  if (req.method !== 'POST' || req.url !== '/send') {
    res.writeHead(404).end('Use POST /send or connect a WebSocket to /ws\n');
    return;
  }
  let body = '';
  req.on('data', (chunk: Buffer) => {
    body += chunk.toString();
  });
  req.on('end', () => {
    const delivered = broadcast(body);
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ delivered }));
  });
});

const wss = new WebSocketServer({ server, path: '/ws' });

wss.on('connection', (socket, req) => {
  console.log(`client connected (${req.socket.remoteAddress}); ${wss.clients.size} total`);
  socket.on('message', (data) => broadcast(data.toString(), socket));
  socket.on('close', () => console.log(`client disconnected; ${wss.clients.size} total`));
});

function broadcast(message: string, sender?: WebSocket): number {
  let delivered = 0;
  for (const client of wss.clients) {
    if (client !== sender && client.readyState === client.OPEN) {
      client.send(message);
      delivered++;
    }
  }
  console.log(`-> ${delivered} client(s): ${message}`);
  return delivered;
}

server.listen(PORT, () => {
  console.log(
    `Postman relay listening on ws://localhost:${PORT}/ws and http://localhost:${PORT}/send`,
  );
});
