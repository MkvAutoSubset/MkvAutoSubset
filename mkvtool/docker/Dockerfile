FROM alpine

RUN apk update && \
    apk add py3-fonttools mkvtoolnix ripgrep ffmpeg git cmake make gcc libc-dev libpng-dev freetype-dev

RUN wget https://api.github.com/repos/MkvAutoSubset/MkvAutoSubset/releases/latest && \
    VERSION=$(grep tag_name latest | cut -d '"' -f 4 | cut -d 'v' -f 2) && \
    rm latest && \
    wget https://github.com/MkvAutoSubset/MkvAutoSubset/releases/download/v${VERSION}/mkvtool_${VERSION}_Linux_$(uname -m).tar.gz && \
    tar -xzvf *.tar.gz && \
    rm *.tar.gz && \
    mv mkvtool /usr/local/bin/ && \
    mkdir fonts

RUN wget https://github.com/libass/libass/releases/download/0.16.0/libass-0.16.0.tar.gz && \
    tar -xzvf libass-0.16.0.tar.gz && \
    cd libass-0.16.0 && \
    ./configure --prefix=/usr/local && \
    make && \
    make install && \
    cd .. && \
    rm -rf libass-0.16.0*

RUN git clone https://github.com/Masaiki/ass2bdnxml.git && \
    cd ass2bdnxml && \
    cmake -Bbuild -DCMAKE_BUILD_TYPE=Release . && \
    cmake --build build && \
    cp build/ass2bdnxml /usr/local/bin/ && \
    cd ~ && \
    rm -rf ass2bdnxml

WORKDIR root
CMD ["sh", "-c", "[ -f ~/.mkvtool/caches/*.cache ] || mkvtool -cc -s /fonts ; sh"]