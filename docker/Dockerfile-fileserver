FROM alpine:3.13

WORKDIR /app

ADD ./build/fileserver/app .
ADD ./config.env .
RUN mkdir storage

CMD ["./app"]