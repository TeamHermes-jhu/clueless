# clueless

A web-based version of Clue

## Backend

The backend is written primarily in Golang.

All go code is stored under ./backend/

### Backend Structure

./backend/cmd/ -> contains the cli commands and flags
associated with the backend server.
./backend/internal/ -> Where internal (not publically exported) code goes.
./backend/internal/gen/ -> generated code, where protos generate to.

### Linting

Find all used go linters under ./backend/.golangci.yml

Linters can be run on the backend by running the command:
`make lint`

### Testing

Testing the backend can be done by running:
`make test`

### Building/Running the Backend

The backend can be built using the command:
`make server`

Then run using
`./backend/server`
