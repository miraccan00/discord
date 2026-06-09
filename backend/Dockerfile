# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server

# Stage 2: minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/bin/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
