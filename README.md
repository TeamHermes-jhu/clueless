# clueless

A web-based version of Clue

## Setting Up

Install the following:

- golang 1.26.4 (backend)
- golangci-lint 2.14.0 (backend linting)
- buf 1.73.0 (protobuf)
- make (Makefile)

## Linting

Linting for the entire repo can be done with:
`make lint`

Backend linting is definde under
`backend/.golangci.yml`

## Testing

Testing for the entire repo can be done with:
`make test`

## Github Workflows

Defined under
`.github/workflows/`

This has all the CICD pipelines that run for this project.

Currently we have:

- lint.yml (runs all linters)

## Backend

The backend is written primarily in Golang.

All go code is stored under ./backend/

### Backend Structure

`./backend/cmd/` -> contains the cli commands and flags
associated with the backend server.
`./backend/internal/` -> Where internal (not publically exported) code goes.
`./backend/internal/gen/` -> generated code, where protos generate to.

### Building/Running the Backend

The backend can be built using the command:
`make server`

Then run using
`./backend/bin/server`
