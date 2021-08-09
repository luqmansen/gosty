FROM alpine:3.12

RUN    buildDeps="build-base \
       zlib-dev \
       freetype-dev \
       jpeg-dev \
       git \
       libmad-dev \
       ffmpeg-dev \
       coreutils \
       yasm-dev \
       lame-dev \
       x264-dev \
       libvpx-dev \
       x265-dev \
       libass-dev \
       libwebp-dev \
       opus-dev \
       libogg-dev \
       libvorbis-dev \
       libtheora-dev \
       libxv-dev \
       alsa-lib-dev \
       xvidcore-dev \
       openssl-dev \
       libpng-dev \
       jack-dev \
       sdl-dev \
       openjpeg-dev \
       expat-dev" && \
       apk  add --no-cache --update ${buildDeps} ffmpeg libxslt openssl libpng bash exiv2 && \
       rm -rf !$/.git && \
       git clone --depth 1 --branch v1.0.1 https://github.com/gpac/gpac.git /tmp/gpac && \
       cd /tmp/gpac && \
       ./configure --static-mp4box --use-zlib=no && \
       JOB=$((2*$(nproc))) && \
       make -j $JOB && \
       make install && \
       make distclean && \
       cd && \
       rm -rf /tmp/gpac && \
       apk del ${buildDeps} && \
       rm -rf /var/cache/apk/*

CMD ["/bin/sh"]