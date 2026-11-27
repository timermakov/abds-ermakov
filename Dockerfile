FROM golang:1.27.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM scratch

COPY --from=build /bin/server /bin/server

ENTRYPOINT ["/bin/server"]
