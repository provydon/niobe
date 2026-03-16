# Go agent for Niobe. Config via env (DB_*, PORT, APP_URL, GEMINI_API_KEY, etc.).
# Build: docker build -f Dockerfile -t niobe-agent .
# Cloud Run sets PORT; DB is via Cloud SQL connector (/cloudsql/INSTANCE).

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /agent .

# PORT default for local; Cloud Run sets PORT
ENV PORT=9000
EXPOSE 9000

CMD ["./agent"]
