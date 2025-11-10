FROM golang:1.25.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p data

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN /go/bin/swag init -g cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/server /app/server
COPY scripts/migrations /app/scripts/migrations
COPY --from=builder /app/docs /app/docs
COPY --from=builder --chown=nonroot:nonroot /app/data /app/data

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/app/server"]

