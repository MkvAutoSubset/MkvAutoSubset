FROM alpine

RUN apk update && \
    apk add py3-fonttools mkvtoolnix ripgrep ffmpeg git cmake make gcc libc-dev libass-dev libpng-dev

RUN wget https://api.github.com/repos/MkvAutoSubset/MkvAutoSubset/releases/latest && \
    VERSION=$(grep tag_name latest | cut -d '"' -f 4 | cut -d 'v' -f 2) && \
    rm latest && \
    wget https://github.com/MkvAutoSubset/MkvAutoSubset/releases/download/v${VERSION}/mkvtool_${VERSION}_Linux_$(uname -m).tar.gz && \
    tar -xzvf *.tar.gz && \
    rm *.tar.gz && \
    mv mkvtool /usr/local/bin/ && \
    mkdir fonts

RUN git clone https://github.com/Masaiki/ass2bdnxml.git && \
    cd ass2bdnxml && \
    cmake -Bbuild -DCMAKE_BUILD_TYPE=Release . && \
    cmake --build build && \
    cp build/libass2bdnxml /usr/local/bin/ && \
    cd ~ && \
    rm -rf ass2bdnxml

WORKDIR root
CMD ["sh", "-c", "[ -f ~/.mkvtool/caches/*.cache ] || mkvtool -cc -s /fonts ; sh"]