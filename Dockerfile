FROM alpine:latest
RUN apk --no-cache add tzdata ca-certificates

FROM scratch

COPY --from=0 /etc/ssl/certs /etc/ssl/certs
COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
ADD ./bin/app /app

EXPOSE 50051
ENTRYPOINT ["/app"]
