# clueless

A web-based version of Clue

## Setting Up

Install the following:

- golang 1.26.4 (backend)
- golangci-lint 2.14.0 (backend linting)
- buf 1.73.0 (protobuf)
- make (Makefile)
- docker

## Linting

Linting for the entire repo can be done with: `make lint`

Backend linting is definde under `backend/.golangci.yml`

## Testing

Testing for the entire repo can be done with: `make test`

## Github Workflows

Defined under `.github/workflows/`

This has all the CICD pipelines that run for this project.

Currently we have:

- lint.yml (runs all linters)

## Backend

The backend is written primarily in Golang.

All go code is stored under ./backend/

## Frontend

The frontend is written in angular.

All angular code is stored under ./frontend/

## Running the app

App can be simply by running `make app`.
