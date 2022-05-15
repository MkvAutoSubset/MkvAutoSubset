NAME="mkvtool"
[ "$(docker images | grep ${NAME})" ] && docker rm ${NAME}
docker build -t ${NAME} $(dirname "$0")
