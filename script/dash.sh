args=""
space=" "
for arg in "$@"
do
    args=$args$space$arg
done

MP4Box -dash 20000 -rap -frag-rap -bs-switching no -profile dashavc264:live -out $args