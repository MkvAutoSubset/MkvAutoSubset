#!/usr/bin/env sh

ARCH_LIST="x86_64 aarch64"
OS_LIST="windows linux macos"

for ARCH in $ARCH_LIST; do
  for OS in $OS_LIST; do
    ZIG_TARGET="${ARCH}-${OS}"

    BUILD_ROOT="/deps/${ZIG_TARGET}"

    ZIG="zig build -Dtarget=${ZIG_TARGET} -Doptimize=ReleaseSmall --prefix ${BUILD_ROOT} --build-file"

    ${ZIG} /package/fribidi-*/build.zig
    ${ZIG} /package/zlib-*/build.zig
    ${ZIG} /package/libpng-*/build.zig -Di=../../deps/${ZIG_TARGET}/include
    ${ZIG} /package/freetype-*/build.zig
    ${ZIG} /package/harfbuzz-*/build.zig
    ${ZIG} /package/libass-*/build.zig -Di=../../deps/${ZIG_TARGET}/include
  done
done

