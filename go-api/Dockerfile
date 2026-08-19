FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget \
	&& adduser -D -H -u 10001 appuser

COPY --from=build /out/server /usr/local/bin/server

USER appuser
EXPOSE 8080
ENV PORT=8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
	CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

CMD ["server"]
