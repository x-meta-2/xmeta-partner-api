# Stage 1: BUILD
FROM golang:1.25-alpine AS build

ENV TZ=Asia/Ulaanbaatar
ENV GO111MODULE=on

RUN apk add --no-cache bash ca-certificates git gcc g++ libc-dev make tzdata
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Source code
COPY . .

# Build steps
RUN go mod tidy
RUN swag init --parseDependency --parseInternal
RUN CGO_ENABLED=0 GOOS=linux go build -a -gcflags='-N -l' -installsuffix cgo -o main .
# Stage 2: Runtime
FROM alpine:latest

ENV TZ=Asia/Ulaanbaatar

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

COPY --from=build /app/main /app/

EXPOSE 8080

CMD ["./main"]
