FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
RUN adduser -D -u 1000 zensor
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o server cmd/api/main.go

FROM scratch
# FROM ubuntu:20.04
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /app/server /server
USER 1000
ENTRYPOINT [ "/server" ]
