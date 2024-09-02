# Mkv Auto Subset

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/MkvAutoSubset/MkvAutoSubset?include_prereleases)

ASS字幕字体子集化 ASS转PGS MKV批量提取/生成

[问题反馈](https://bbs.acgrip.com/thread-9070-1-1.html)

## 什么是字幕字体子集化
- 这里说的字幕特指ass(ssa)这种带有特效的文本字幕;
- ass字幕会引用一些字体,这些字体在播放器所在的系统里可能有安装,也可能没有;
- 为了实现在任意地方都能有完整的视觉体验,可以把字幕以及字幕里所引用的字体文件一起打包进mkv文件里;
- 以上的操作存在一个问题,有些字幕会引用很多字体,这些字体文件体积动辄几十MB,而字幕只用到了其中的几个字而已;
- 比如一个番剧本体200M,但打包了字体文件后变成400M了,这像画吗?
- 综上所述,子集化的目的就是把字体拆包,找出字幕用到的那部分字形并重新打包;
- 好处不仅限于节约存储空间,加快缓冲速度;
- 想想看:在一个只有30Mbps上传的网络环境下,要看上面那个光字体就200M的番剧,这河里吗?

## next 分支
- 用C重新实现了字幕处理功能
- 不再依赖基于Python的FontTools工具进行子集化
- 不再依赖ass2bdnxml工具进行ASS转PGS
- 不再强制需要其他的工具依赖:
  - MKVToolNix:涉及抽取/混流操作时,依然需要
  - FFmpeg:涉及生成测试视频时,依然需要
- 丧失了方便地跨平台编译的能力
- 目前只自动化提供Windows,macOS,Linux三个平台的amd64及arm64架构的执行文件
- 其他平台请自行编译

### Docker镜像使用说明
- 从Dockerhub获取
  ```shell
  FONT_DIR="/usr/share/fonts/truetype" #字体目录
  CACHE_DIR="${HOME}/.mkvtool/caches"  #缓存目录
  OTHER_DIR="" #其他目录(可选,示例见下节.)
  docker pull ac79b0c6/mkvtool #拉取/更新镜像
  docker run --rm -it -v ${FONT_DIR}:/fonts -v ${CACHE_DIR}:/root/.mkvtool/caches ${OTHER_DIR} ac79b0c6/mkvtool:${TAGNAME} #运行镜像
  ```
- 手动构建&运行
  ```shell
  git clone https://github.com/MkvAutoSubset/MkvAutoSubset.git #克隆项目
  cd MkvAutoSubset #进入项目目录
  sh docker/rebuild.sh #构建镜像
  cp docker/run.sh docker/run_my.sh  #拷贝一份自己的运行脚本
  vi docker/run_my.sh #修改自己的运行脚本(可选)
  sh docker/run_my.sh #运行镜像
  ```
- docker/run_my.sh的修改说明
  * FONT_DIR: 字体文件目录
  * CACHE_DIR: 缓存目录
  * OTHER_DIR: 其他目录(可选)
    * 示例:“-v ${HOME}/work:/work”

### 通过Docker镜像编译
  - 从Dockerhub获取:
  ```shell
  TAGNAME=next #使用next分支的镜像
  DIST_DIR="${HOME}/mkvtool_dist"  #编译结果目录
  docker pull ac79b0c6/mkvtool-builder:${TAGNAME} #拉取/更新镜像
  docker run --rm -it -v ${DIST_DIR}:/dist ac79b0c6/mkvtool-builder:${TAGNAME} #运行镜像
  ```
  - 手动构建&运行:
  ```shell
  git clone https://github.com/MkvAutoSubset/MkvAutoSubset.git #克隆项目
  cd MkvAutoSubset #进入项目目录
  sh builder-docker/run.sh #构建并运行镜像
  ```

### 通过根目录的脚本编译
- 请确保已经安装了**go**,**gcc**,**vcpkg**,**dotnet**
- 以上项目都在PATH环境变量里
- 将vcpkg的路径添加到环境变量 **VCPKG_ROOT**
- 运行根目录的编译脚本
- 成品在build目录下
- #### 成品说明:
  - __mkvtool-cli__:命令行版本可执行文件
  - __mkvtool-gui__:GUI版本可执行文件
  - __mkvtool.so__:可用于二次开发的动态链接库 [SDK调用文档](mkvlib/sdk/README.md)
  - 注意:__mkvtool-gui__ 需要和 __mkvtool.so__ 在同一目录下才能正常运行,另外 __libHarfBuzzSharp__ 和 __libSkiaSharp__ 也是 __mkvtool-gui__ 的依赖库,要使用GUI版本这些文件也需要在同一目录下.

### 手动编译过程(以Linux为例)
- 前置操作同上
- ```shell
  cd MkvAutoSubset/mkvtool
  go mod tidy
  VCPKG_ROOT="${HOME}/vcpkg" #你的vcpkg路径
  export VCPKG_DEFAULT_TRIPLET="x64-linux-release" #你的vcpkg triplet三元组
  export VCPKG_BUILD_TYPE="Release"
  ${VCPKG_ROOT}/vcpkg install fribidi freetype[core,zlib,png] harfbuzz[core,experimental-api]
  ${VCPKG_ROOT}/vcpkg install libass #安装依赖
  PATH_ROOT="${VCPKG_ROOT}/installed/${VCPKG_DEFAULT_TRIPLET}"
  H_PATH="${PATH_ROOT}/include"
  L_PATH="${PATH_ROOT}/lib"
  export CGO_CFLAGS="-I${H_PATH} -DHB_EXPERIMENTAL_API -Os"
  export CGO_LDFLAGS="-L${L_PATH} -static -lass -lfreetype -lm -lz -lfontconfig -lpng -lfribidi -lharfbuzz -lharfbuzz-subset -lexpat"
  go build #编译
  ```

### 其他依赖 - 按需安装

- MKVToolNix - 用于提取/混流字幕和字体
  ```shell
  apt install mkvtoolnix #Debian/Ubuntu
  apk add mkvtoolnix #Alpine
  brew install mkvtoolnix #macOS
  ```
- FFmpeg - 用于生成测试视频/烧录字幕
  ```shell
  apt install ffmpeg #Debian/Ubuntu
  apk add ffmpeg #Alpine
  brew install ffmpeg #macOS
  ```

#### 关于Windows用户

- 从 [这里](https://www.fosshub.com/MKVToolNix.html) 下载并安装MKVToolNix
- 从 [这里](https://ffmpeg.org/download.html) 下载并安装ffmpeg
- 保证以上两个依赖项的相关可执行文件(_mkvextract.exe_,_mkvmerge.exe_,_ffmpeg.exe_)在 **path** 环境变量里

### 使用说明同[主线](https://github.com/MkvAutoSubset/MkvAutoSubset/tree/master?tab=readme-ov-file#%E4%B8%80%E9%83%A8%E5%88%86%E4%B8%AD%E6%96%87%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E%E8%8B%B1%E6%96%87%E5%AE%8C%E6%95%B4%E7%89%88)