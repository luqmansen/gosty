ffmpeg -y -f mpegts -i "concat:$1" -c copy -bsf:a aac_adtstoasc $2
