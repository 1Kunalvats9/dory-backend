# --- Build Stage ---
FROM golang:1.25-alpine AS builder

# Set the working directory
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o dory-server cmd/api/main.go

# --- STEP 2: The Final Tiny Image --
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/dory-server .

# Render uses the PORT environment variable. 
# Your Go app uses config.AppConfig.Port, make sure it defaults to the PORT env var.
EXPOSE 8080

CMD ["./dory-server"]