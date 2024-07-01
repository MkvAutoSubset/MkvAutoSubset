FROM debian:stable-slim as builder

ENV VCPKG_ROOT=/vcpkg
ENV VCPKG_DEFAULT_TRIPLET="x64-linux-release"
ENV VCPKG_BUILD_TYPE="Release"

RUN apt-get update && \
    apt-get install -y curl zip unzip tar git build-essential pkg-config python3 && \
    git clone https://github.com/microsoft/vcpkg && \
    ${VCPKG_ROOT}/bootstrap-vcpkg.sh -disableMetrics && \
    ${VCPKG_ROOT}/vcpkg install fribidi freetype[core,zlib,png] harfbuzz[core,experimental-api] && \
    ${VCPKG_ROOT}/vcpkg install libass

RUN curl -O https://dl.google.com/go/go1.22.4.linux-amd64.tar.gz && \
    tar zxf go*
ENV PATH=${PATH}:/go/bin

RUN git clone https://github.com/MkvAutoSubset/MkvAutoSubset -b next && \
    cd /MkvAutoSubset/mkvtool && \
    go mod tidy && \
    PATH_ROOT="${VCPKG_ROOT}/installed/${VCPKG_DEFAULT_TRIPLET}" && \
    H_PATH="${PATH_ROOT}/include" && \
    L_PATH="${PATH_ROOT}/lib" && \
    CGO_CFLAGS="-I${H_PATH} -DHB_EXPERIMENTAL_API -Os" && \
    CGO_LDFLAGS="-L${L_PATH} -static -lass -lfreetype -lm -lz -lfontconfig -lpng -lfribidi -lharfbuzz -lharfbuzz-subset -lexpat" && \
    CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} go build -ldflags "-s -w"

FROM alpine

COPY --from=builder /MkvAutoSubset/mkvtool/mkvtool /usr/bin
RUN apk add ffmpeg mkvtoolnix && \
    mkdir fonts work

WORKDIR work
CMD ["sh", "-c", "[ -f ~/.mkvtool/caches/*.cache ] || mkvtool cache /fonts ; sh"]