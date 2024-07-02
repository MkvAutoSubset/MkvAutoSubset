#!/usr/bin sh

cd $(dirname "$0")
rm -rf build > /dev/null 2>&1
mkdir build

VCPKG_DEFAULT_TRIPLET="x64-linux-release"
VCPKG_BUILD_TYPE="Release"
vcpkg install  fribidi freetype[core,zlib,png] harfbuzz[core,experimental-api]
vcpkg install libass

PATH_ROOT="${VCPKG_ROOT}/installed/${VCPKG_DEFAULT_TRIPLET}"
H_PATH="${PATH_ROOT}/include"
L_PATH="${PATH_ROOT}/lib"
CGO_CFLAGS="-I${H_PATH} -DHB_EXPERIMENTAL_API -Os"
CGO_LDFLAGS="-L${L_PATH} -static -lass -lfreetype -lm -lz -lfontconfig -lpng -lfribidi -lharfbuzz -lharfbuzz-subset -lexpat"

LDFLAGS="-s -w"

cd mkvtool
go mod tidy
CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} go build -ldflags "${LDFLAGS}" -o ../build/mkvtool-cli
cd ..

cd mkvlib/sdk
go mod tidy
CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} go build -ldflags "${LDFLAGS}" -buildmode c-shared -o ../../build/mkvlib.so && rm ../../build/mkvlib.h
cd ../..

cd mkvtool-gui
dotnet publish --sc -c Release /p:PublishSingleFile=true -p:AssemblyName=mkvtool-gui -p:DebugType=none -o ../build
cd ..