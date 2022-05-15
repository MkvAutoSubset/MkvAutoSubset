docker rm mkvtool
docker rmi mkvtool
docker build -t ${NAME} $(dirname "$0")