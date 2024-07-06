NAME="mkvtool"
[ "$(docker images | grep ${NAME})" ] && docker rmi ${NAME}
docker build -t ${NAME} $(dirname "$0")
