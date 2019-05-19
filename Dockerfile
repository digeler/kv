FROM ubuntu:16.04
RUN apt-get update
RUN apt-get install ca-certificates -y
RUN rm -rf /var/cache/apk/*
COPY kv /app
