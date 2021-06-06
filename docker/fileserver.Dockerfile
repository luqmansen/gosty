FROM rastasheep/ubuntu-sshd:18.04

RUN apt update && apt install sshpass lsyncd -y

WORKDIR /app

ADD ./build/fileserver/fileserver .
ADD ./config.env .
ADD script/lsyncd.sh ./script/lsyncd.sh
RUN mkdir storage

EXPOSE 8001 22
CMD ["/bin/sh","-c","/usr/sbin/sshd && ./fileserver"]
