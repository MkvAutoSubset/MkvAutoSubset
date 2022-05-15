FROM alpine

RUN apk update && \
    apk add py3-fonttools mkvtoolnix

ENV VERSION 4.0.7

RUN wget https://github.com/MkvAutoSubset/MkvAutoSubset/releases/download/v${VERSION}/mkvtool_${VERSION}_Linux_$(uname -m).tar.gz && \
    tar -xzvf mkvtool_${VERSION}_Linux_$(uname -m).tar.gz && \
    rm mkvtool_${VERSION}_Linux_$(uname -m).tar.gz && \
    mv mkvtool /usr/local/bin/ && \
    mkdir fonts

WORKDIR ~
CMD ["sh", "-c", "[ -d ~/.mkvtool/caches ] || mkvtool -cc -s /fonts ; sh"]