FROM alpine

RUN apk update && \
    apk add py3-fonttools mkvtoolnix ripgrep ffmpeg

RUN wget https://api.github.com/repos/MkvAutoSubset/MkvAutoSubset/releases/latest && \
    VERSION=$(grep tag_name latest | cut -d '"' -f 4 | cut -d 'v' -f 2) && \
    rm latest && \
    wget https://github.com/MkvAutoSubset/MkvAutoSubset/releases/download/v${VERSION}/mkvtool_${VERSION}_Linux_$(uname -m).tar.gz && \
    tar -xzvf *.tar.gz && \
    rm *.tar.gz && \
    mv mkvtool /usr/local/bin/ && \
    mkdir fonts \

WORKDIR root
CMD ["sh", "-c", "[ -f ~/.mkvtool/caches/*.cache ] || mkvtool -cc -s /fonts ; sh"]