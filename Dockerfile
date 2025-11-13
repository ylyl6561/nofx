FROM golang:1.25.0-alpine AS builder

LABEL "language"="go"

RUN apk add --no-cache nodejs npm

WORKDIR /src

COPY . .

RUN go mod download

RUN cd web && npm install && npm run build && cd ..

RUN go build -o app main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /src/app .

EXPOSE 8080

CMD ["./app"]