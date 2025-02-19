FROM alpine:3.8 as build
RUN adduser -D -u 1000 zensor
WORKDIR /app
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o server cmd/api/main.go

FROM scratch
# FROM ubuntu:20.04
COPY --from=build /app/server /server
COPY --from=build /etc/passwd /etc/passwd
USER 1000
ENTRYPOINT [ "/server" ]
