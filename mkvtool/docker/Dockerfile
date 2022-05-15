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
    mkdir fonts

RUN apk add cmake make gcc
RUN apk add libc-dev libpng-dev freetype-dev fribidi-dev harfbuzz-dev gettext-dev

RUN wget https://api.github.com/repos/libass/libass/releases/latest && \
    VERSION=$(grep tag_name latest | cut -d '"' -f 4) && \
    wget https://github.com/libass/libass/releases/download/${VERSION}/libass-${VERSION}.tar.gz && \
    rm latest && \
    tar -xzvf *.tar.gz && \
    cd libass* && \
    ./configure && \
    make install && \
    cd .. && \
    rm -rf libass*

RUN wget https://api.github.com/repos/Masaiki/ass2bdnxml/releases/latest && \
    VERSION=$(grep tag_name latest | cut -d '"' -f 4) && \
    rm latest && \
    wget https://github.com/Masaiki/ass2bdnxml/archive/refs/tags/${VERSION}.tar.gz && \
    tar -xzvf *.tar.gz && \
    cd ass2bdnxml* && \
    cmake -Bbuild -DCMAKE_BUILD_TYPE=Release . && \
    cmake --build build && \
    cp build/ass2bdnxml /usr/local/bin/ && \
    cd .. && \
    rm -rf ass2bdnxml*

WORKDIR root
CMD ["sh", "-c", "[ -f ~/.mkvtool/caches/*.cache ] || mkvtool -cc -s /fonts ; sh"]