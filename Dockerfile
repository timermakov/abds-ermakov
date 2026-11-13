FROM golang:1.27.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -mod=vendor -o /bin/server ./cmd/server

FROM scratch

COPY --from=build /bin/server /bin/server

ENTRYPOINT ["/bin/server"]
