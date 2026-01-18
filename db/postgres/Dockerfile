FROM postgres:16-alpine

ENV TZ=Asia/Tokyo

RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

# COPY ./init.sql /docker-entrypoint-initdb.d/