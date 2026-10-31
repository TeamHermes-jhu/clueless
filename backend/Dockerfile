FROM golang:1.26.4 AS build-stage

WORKDIR /app

COPY go.mod  go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /server

FROM gcr.io/distroless/base-debian12 AS release-stage

WORKDIR /

COPY --from=build-stage /server /server

EXPOSE 8000

USER nonroot:nonroot

ENTRYPOINT [ "/server", "serve" ]
