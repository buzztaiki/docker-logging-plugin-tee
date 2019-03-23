# Docker Log Driver Tee

A docker log driver that send logs to multiple log drivers.

WIP.

## Example

```
> docker run --rm --log-driver buzztaiki/docker-log-driver-tee:development \
	--log-opt drivers=json-file,syslog \
    --log-opt syslog:syslog-address=tcp://172.17.0.1:1514 \
    --log-opt syslog:syslog-format=rfc5424 \
    ubuntu:bionic sh -c 'while :; do date; sleep 2; done'
```

## License

MIT
