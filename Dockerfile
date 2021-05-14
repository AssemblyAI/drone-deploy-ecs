FROM alpine

RUN apk add -Uuv ca-certificates

COPY bin/deploy /opt/

CMD ["/opt/deploy"]