FROM luqmansen/alpine-ffmpeg-mp4box

WORKDIR /app

ADD ./build/worker/app ./app
ADD script/*.sh ./script/
ADD ./config.env .
RUN mkdir tmp-worker

CMD ["./app"]