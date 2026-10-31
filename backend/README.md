# Backend Server

## Structure

`./backend/cmd/` -> contains the cli commands and flags associated with the
backend server. `./backend/internal/` -> Where internal (not publically
exported) code goes. `./backend/internal/gen/` -> generated code, where protos
generate to.

## Building/Running the server

The server can be built using the command: `make server`

Then run using `./backend/server`
