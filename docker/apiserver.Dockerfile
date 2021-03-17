FROM jrottenberg/ffmpeg:4.3-alpine312

WORKDIR /app

# uncomment this to check when build context get bigger
#COPY . /tmp/build
#RUN find /tmp/build

#docker alpine by default doesn't have mime.types file
ADD "http://svn.apache.org/viewvc/httpd/httpd/trunk/docs/conf/mime.types?revision=1884511&view=co" /etc/mime.types
COPY ./build/apiserver/app .
COPY ./config.env .
RUN mkdir tmp

#to override base image entrypoint
ENTRYPOINT ["/usr/bin/env"]

CMD ["./app"]