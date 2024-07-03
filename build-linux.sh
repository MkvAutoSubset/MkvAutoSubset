#!/usr/bin/env sh

cd $(dirname "$0")
rm -rf build > /dev/null 2>&1
mkdir build

export VCPKG_DEFAULT_TRIPLET="x64-linux-release"
export VCPKG_BUILD_TYPE="Release"
${VCPKG_ROOT}/vcpkg install  fribidi freetype[core,zlib,png] harfbuzz[core,experimental-api]
${VCPKG_ROOT}/vcpkg install libass

export PATH_ROOT="${VCPKG_ROOT}/installed/${VCPKG_DEFAULT_TRIPLET}"
export H_PATH="${PATH_ROOT}/include"
export L_PATH="${PATH_ROOT}/lib"
export CGO_CFLAGS="-I${H_PATH} -DHB_EXPERIMENTAL_API -Os"
export CGO_LDFLAGS="-L${L_PATH} -static -lass -lfreetype -lm -lz -lfontconfig -lpng -lfribidi -lharfbuzz -lharfbuzz-subset -lexpat"

export LDFLAGS="-s -w"

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