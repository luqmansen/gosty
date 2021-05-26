FROM debian:stable-slim

WORKDIR /app

ADD ./build/fileserver/fileserver .
ADD ./config.env .
RUN mkdir storage

EXPOSE 8001
CMD ["./fileserver"]