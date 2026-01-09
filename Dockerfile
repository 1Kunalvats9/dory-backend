# --- STEP 1: The Build Environment ---
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
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

EXPOSE 3000

CMD ["./dory-server"]