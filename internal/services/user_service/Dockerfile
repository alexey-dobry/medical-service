FROM golang:1.26-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local

COPY internal/pkg/go.mod internal/pkg/go.sum internal/pkg/
COPY internal/services/user_service/go.mod internal/services/user_service/go.sum internal/services/user_service/
RUN cd internal/services/user_service && go mod download

COPY internal/pkg internal/pkg
COPY internal/services/user_service internal/services/user_service
RUN cd internal/services/user_service && go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app \
    && mkdir -p /app/logs /app/configs \
    && chown -R app /app
WORKDIR /app
COPY --from=build /out/service /app/service
USER app
ENTRYPOINT ["/app/service", "-config", "/app/configs/config.yaml"]