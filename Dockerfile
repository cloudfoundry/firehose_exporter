FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Ferran Rodenas <frodenas@gmail.com>

COPY firehose_exporter /bin/firehose_exporter

EXPOSE      9186
ENTRYPOINT  [ "/bin/firehose_exporter" ]
