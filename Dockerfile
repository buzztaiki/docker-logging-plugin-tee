FROM scratch
COPY ./docker-log-driver-tee /usr/bin/
CMD ["/usr/bin/docker-log-driver-tee"]
