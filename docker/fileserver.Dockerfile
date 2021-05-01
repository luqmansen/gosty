FROM alpine:3.13

WORKDIR /app

ADD ./build/fileserver/fileserver .
ADD ./config.env .
RUN mkdir storage

CMD ["./fileserver"]