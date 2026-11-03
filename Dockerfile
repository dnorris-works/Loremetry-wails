FROM node:20-bookworm AS web
WORKDIR /web
COPY website/package.json website/package-lock.json ./
RUN npm ci
COPY website/ ./
RUN npm run build

FROM golang:1.25-bookworm AS api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/* \
 && useradd -u 1000 -m miget
WORKDIR /app
COPY --from=api --chown=1000:1000 /out/server /app/server
COPY --from=web --chown=1000:1000 /web/dist /app/website
USER 1000
ENV PORT=5000
ENV WEBSITE_DIR=/app/website
EXPOSE 5000
CMD ["./server"]
