FROM golang:1.23-bullseye AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash gcc make git libc6-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/api ./cmd/api

# ===== STAGE 2: Minimal runtime =====
FROM debian:bullseye-slim AS api

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libssl1.1 libc6 libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/api /api

CMD ["/api"]
