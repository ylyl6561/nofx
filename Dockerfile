FROM node:22-alpine AS frontend-builder
WORKDIR /src
COPY web /src/web
WORKDIR /src/web
RUN npm install && npm run build

FROM golang:1.25.0-alpine AS backend-builder
WORKDIR /src
COPY . /src
COPY --from=frontend-builder /src/web/dist /src/web/dist
RUN go mod download
RUN go build -o app main.go

FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /src/app /app/app
COPY --from=backend-builder /src/web/dist /app/web/dist
EXPOSE 8080
CMD ["/app/app"]
