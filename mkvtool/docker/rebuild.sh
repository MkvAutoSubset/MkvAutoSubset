NAME='mkvtool'
docker rm ${NAME}
docker rmi ${NAME}
docker build -t ${NAME} $(dirname "$0")