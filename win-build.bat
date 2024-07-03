@echo off

cd /d %~dp0
rd /s/q build > NUL 2>&1
md build

set VCPKG_DEFAULT_TRIPLET=x64-mingw-static
set VCPKG_BUILD_TYPE=Release
%VCPKG_ROOT%\vcpkg install fribidi freetype[core,zlib,png] harfbuzz[core,experimental-api]
%VCPKG_ROOT%\vcpkg install libass

set PATH_ROOT=%VCPKG_ROOT%/installed/%VCPKG_DEFAULT_TRIPLET%
set H_PATH=%PATH_ROOT%/include
set L_PATH=%PATH_ROOT%/lib
set CGO_CFLAGS=-I%H_PATH% -DHB_EXPERIMENTAL_API -Os
set CGO_LDFLAGS=-L%L_PATH% -lharfbuzz-subset -lass -lpng -lfreetype -lharfbuzz -lfribidi -lzlib -lgdi32

set LDFLAGS=-s -w

cd mkvtool
go mod tidy
go build -ldflags="%LDFLAGS%" -o ..\build\mkvtool-cli.exe
cd ..

cd mkvlib\sdk
go mod tidy
go build -ldflags="%LDFLAGS%" -buildmode c-shared -o ..\..\build\mkvlib.so && del ..\..\build\mkvlib.h
cd ..\..

cd mkvtool-gui
dotnet publish --sc -c Release /p:PublishSingleFile=true -p:AssemblyName=mkvtool-gui -p:DebugType=none -o ..\build
cd ..