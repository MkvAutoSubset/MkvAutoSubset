#!/usr/bin/env bash

cd $(dirname "$0")
rm -rf build > /dev/null 2>&1
mkdir build

if [ -z "${VCPKG_ROOT}" ]; then
   VCPKG_ROOT="${HOME}/vcpkg"
fi

if [ ! -d "${VCPKG_ROOT}" ]; then
   git clone https://github.com/microsoft/vcpkg ${VCPKG_ROOT}
   ${VCPKG_ROOT}/bootstrap-vcpkg.sh -disableMetrics
fi

if ! which pkg-config &> /dev/null; then
    if ! which brew &> /dev/null; then
        echo "Error: brew is not installed." >&2
        exit 1
    else
        brew install pkg-config
    fi
fi

export VCPKG_DEFAULT_TRIPLET="x64-osx-release"
export VCPKG_BUILD_TYPE="Release"
${VCPKG_ROOT}/vcpkg install fribidi "freetype[core,zlib,png]" "harfbuzz[core,experimental-api]"
${VCPKG_ROOT}/vcpkg install libass

PATH_ROOT="${VCPKG_ROOT}/installed/${VCPKG_DEFAULT_TRIPLET}"
H_PATH="${PATH_ROOT}/include"
L_PATH="${PATH_ROOT}/lib"
export CGO_CFLAGS="-I${H_PATH} -DHB_EXPERIMENTAL_API -Os"
export CGO_LDFLAGS="-L${L_PATH} -lass -lfreetype -lz -lpng -lfribidi -lharfbuzz -lharfbuzz-subset -framework CoreText"

LDFLAGS="-s -w"

cd mkvtool
go mod tidy
go build -ldflags "${LDFLAGS}" -o ../build/mkvtool-cli
cd ..

cd mkvlib/sdk
go mod tidy
go build -ldflags "${LDFLAGS}" -buildmode c-shared -o ../../build/mkvlib.so && rm ../../build/mkvlib.h
cd ../..

cd mkvtool-gui
dotnet publish --sc -c Release /p:PublishSingleFile=true -p:AssemblyName=mkvtool-gui -p:DebugType=none -o ../build
cd ..