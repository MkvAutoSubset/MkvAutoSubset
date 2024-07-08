#!/usr/bin/env sh

convert_arch() {
  case $1 in
    amd64)
      echo "x86_64"
      ;;
    arm64)
      echo "aarch64"
      ;;
    *)
      echo "$1"
      ;;
  esac
}

echo "Initializing..."

rm -rf /dist/* > /dev/null

export CGO_ENABLED=1

wget -q -O src.zip https://github.com/MkvAutoSubset/MkvAutoSubset/archive/refs/heads/next.zip > /dev/null
unzip -n -d /package src.zip > /dev/null
rm src.zip

cd /package/MkvAutoSubset-next/mkvtool
go mod tidy > /dev/null 2>&1

ARCH_LIST="amd64 arm64"
OS_LIST="windows linux darwin"

export CGO_ENABLED=1

for ARCH in $ARCH_LIST; do
  for OS in $OS_LIST; do
    export GOOS=$OS
    export GOARCH=$ARCH

    ZIG_TARGET="$(convert_arch $ARCH)-$OS"

    if [ "$OS" = "darwin" ]; then
      OS="osx"
      ZIG_TARGET="$(convert_arch $ARCH)-macos"
      export CC="o64-clang"
    else
      export CC="zig cc -target $ZIG_TARGET"
    fi

    export CGO_CFLAGS="-O3 -I/deps/$ZIG_TARGET/include -DHB_EXPERIMENTAL_API"
    export CGO_LDFLAGS="-L/deps/$ZIG_TARGET/lib -lstdc++ -lfribidi -lzlib -lfreetype -lpng -lharfbuzz -lharfbuzz-subset -lass"

    OUTFILE="/dist/mkvtool-$OS-$GOARCH"
    if [ "$OS" = "windows" ]; then
      OUTFILE="$OUTFILE.exe"
    fi

    clear
    echo "Building $OS-$ARCH..."

    go build -ldflags "-s -w" -o $OUTFILE > /dev/null 2>&1
  done
done

clear
echo "Packaging..."
cd /dist
zip /dist/mkvtool-next.zip * > /dev/null 2>&1
clear
ls -lh /dist
echo "All done."
