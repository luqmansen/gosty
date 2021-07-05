#!/bin/bash

FILE="$1"
SEGMENT_TIME="$2"
BASENAME="${FILE%.*}"
EXTENSION="mp4"

ffmpeg -i $FILE -f segment -segment_time $SEGMENT_TIME -c copy $BASENAME-%d.$EXTENSION
