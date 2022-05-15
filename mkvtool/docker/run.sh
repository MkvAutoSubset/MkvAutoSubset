NAME="mkvtool"
[ $(docker images | grep ${NAME}) ] && docker build -t ${NAME} $(dirname "$0")
FONT_DIR="/usr/share/fonts/truetype" # Change this to your font directory
CACHE_DIR="${HOME}/.mkvtool/caches"  # Change this to your cache directory
OTHER_DIR=""                         # Change this to your other directory for example: -v aaa:bbb
docker run --rm -it -v ${FONT_DIR}:/fonts -v ${CACHE_DIR}:/root/.mkvtool/caches ${OTHER_DIR} ${NAME}
