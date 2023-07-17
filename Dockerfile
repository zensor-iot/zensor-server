FROM alpine:3.8
RUN adduser -D -u 1000 zensor

# FROM scratch
FROM ubuntu:20.04
COPY /server /server
COPY --from=0 /etc/passwd /etc/passwd
USER 1000
ENTRYPOINT [ "/server" ]