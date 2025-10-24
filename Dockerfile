FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=development
ARG COMMIT_HASH=unknown
RUN adduser -D -u 1000 zensor
RUN apk --no-cache add tzdata
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-X zensor-server/internal/infra/node.Version=${VERSION} -X zensor-server/internal/infra/node.CommitHash=${COMMIT_HASH}" \
    -o server cmd/api/main.go

FROM scratch
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /app/server /server
USER 1000
ENTRYPOINT [ "/server" ]
