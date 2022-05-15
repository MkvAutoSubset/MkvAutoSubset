NAME='mkvtool'
docker images | grep ${NAME} || docker build -t ${NAME} $(dirname "0")
FONT_DIR='/usr/share/fonts/truetype' # Change this to your font directory
OTHER_DIR=''                         # Change this to your other directory for example: -v aaa:bbb
docker run --name ${NAME} -it -v ${FONT_DIR}:/fonts ${OTHER_DIR} ${NAME}
