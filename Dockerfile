FROM alpine:3.9 as certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./docker-log-driver-tee /usr/bin/
CMD ["/usr/bin/docker-log-driver-tee"]
