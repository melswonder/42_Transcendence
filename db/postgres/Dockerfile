FROM postgres:16-alpine

RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/timezone

# COPY ./init.sql /docker-entrypoint-initdb.d/