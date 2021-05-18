FROM alpine:3.13

WORKDIR /app

ADD ./build/fileserver/fileserver .
ADD ./config.env .
RUN mkdir storage

EXPOSE 8001
CMD ["./fileserver"]