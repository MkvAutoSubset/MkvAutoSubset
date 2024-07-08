FROM nemirtingas/osxcross_vcpkg:SDK12.1_clang17 as builder

ADD https://go.dev/dl/go1.22.5.linux-amd64.tar.gz /compiler/
ADD https://ziglang.org/download/0.13.0/zig-linux-x86_64-0.13.0.tar.xz /compiler/

RUN cd /compiler && \
    tar zxf go* && \
    rm go*.gz && \
    tar Jxf zig* && \
    rm zig*.xz && \
    mv zig* zig

ENV PATH=${PATH}:/compiler/go/bin:/compiler/zig

ADD https://download.savannah.gnu.org/releases/freetype/freetype-2.13.2.tar.xz /package/
ADD https://github.com/fribidi/fribidi/releases/download/v1.0.15/fribidi-1.0.15.tar.xz /package/
ADD https://github.com/harfbuzz/harfbuzz/releases/download/9.0.0/harfbuzz-9.0.0.tar.xz /package/
ADD https://github.com/libass/libass/releases/download/0.17.3/libass-0.17.3.tar.xz /package/
ADD https://github.com/madler/zlib/releases/download/v1.3.1/zlib-1.3.1.tar.xz /package/
ADD https://github.com/pnggroup/libpng/archive/refs/tags/v1.6.43.tar.gz /package/

RUN cd /package && \
    tar Jxf freetype-* && \
    rm freetype-*.xz && \
    tar Jxf fribidi-* && \
    rm fribidi-*.xz && \
    tar Jxf harfbuzz-* && \
    rm harfbuzz-*.xz && \
    tar Jxf libass-* && \
    rm libass-*.xz && \
    tar Jxf zlib-* && \
    rm zlib-*.xz && \
    tar zxf *.gz && \
    rm *.gz

COPY profile /profile
RUN ln -s /profile/zlib.zig $(ls -d /package/zlib-*)/build.zig && \
    ln -s /profile/png.zig $(ls -d /package/libpng-*)/build.zig && \
    ln -s /profile/freetype.zig $(ls -d /package/freetype-*)/build.zig && \
    ln -s /profile/fribidi.zig $(ls -d /package/fribidi-*)/build.zig && \
    ln -s /profile/harfbuzz.zig $(ls -d /package/harfbuzz-*)/build.zig && \
    ln -s /profile/ass.zig $(ls -d /package/libass-*)/build.zig

COPY include /include
RUN ln -s /include/fribidi-config.h $(ls -d /package/fribidi-*)/lib/ && \
    ln -s /include/pnglibconf.h $(ls -d /package/libpng-*)/ && \
    ln -s /include/config.h $(ls -d /package/libass-*)/

COPY script/bootstrap.sh /script/
RUN chmod +x /script/bootstrap.sh && \
    /script/bootstrap.sh

COPY script/build.sh /script/
RUN chmod +x /script/build.sh

FROM nemirtingas/osxcross_vcpkg:SDK12.1_clang17

COPY --from=builder /deps /
COPY --from=builder /cript/build.sh /

CMD ["/build.sh"]