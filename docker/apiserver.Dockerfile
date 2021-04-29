FROM jrottenberg/ffmpeg:4.3-alpine312

WORKDIR /app

# uncomment this to debug when build context get bigger
#COPY . /tmp/build
#RUN find /tmp/build

#docker alpine by default doesn't have mime.types file
ADD "https://gist.githubusercontent.com/luqmansen/690dd7e79d2f8c7bb9046f5e404ef5c6/raw/86e75df7eef30bafb5fee0162fd9b9a27265ff14/mime.types" /etc/mime.types
COPY ./build/apiserver/apiserver .
COPY ./config.env .
RUN mkdir tmp

#to override base image entrypoint
ENTRYPOINT ["/usr/bin/env"]

CMD ["./app"]