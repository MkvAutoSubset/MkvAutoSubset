# Mkv Auto Subset

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/MkvAutoSubset/MkvAutoSubset?include_prereleases)

ASS字幕字体子集化 MKV批量提取/生成

[问题反馈](https://bbs.acgrip.com/thread-9070-13-1.html)

## 什么叫字幕字体子集化
- 这里说的字幕特指ass(ssa)这种带有特效的文本字幕;
- ass字幕会引用一些字体,这些字体在播放器所在的系统里可能有安装,也可能没有;
- 为了实现在任意地方都能有完整的视觉体验,可以把字幕以及字幕里所引用的字体文件一起打包进mkv文件里;
- 以上的操作存在一个问题,有些字幕会引用很多字体,这些字体文件体积动辄几十MB,而字幕只用到了其中的几个字而已;
- 比如一个番剧本体200M,但打包了字体文件后变成400M了,这像画吗?
- 综上所述,子集化的目的就是把字体拆包,找出字幕用到的那部分字形并重新打包;
- 好处不仅限于节约存储空间,加快缓冲速度;
- 想想看:在一个只有30Mbps上传的网络环境下,要看上面那个光字体就200M的番剧,这河里吗?

## next 分支
用C重新实现了字幕处理功能,但丧失了方便地跨平台编译的能力.如果你有兴趣,可以尝试手动编译并使用.

### 手动编译过程(以Linux为例)
- 配置好go
- 配置好gcc
- 配置好vcpkg
- 克隆本项目
- ```shell
  cd MkvAutoSubset/mkvtool
  go mod tidy
  VCPKG_ROOT=~/vcpkg #你的vcpkg路径
  VCPKG_TRIPLET="x64-linux-release" #你的vcpkg triplet三元组
  vcpkg install --triplet $VCPKG_TRIPLET harfbuzz[experimental-api] libass #安装依赖
  PATH_ROOT="${VCPKG_ROOT}/installed/${VCPKG_TRIPLET}"
  H_PATH="${PATH_ROOT}/include"
  L_PATH="${PATH_ROOT}/lib"
  CGO_CFLAGS="-I${H_PATH} -DHB_EXPERIMENTAL_API -Os"
  CGO_LDFLAGS="-L${L_PATH} -lass -lfreetype -lz -lfontconfig -lpng -lm -lbz2 -lfribidi -lharfbuzz -lharfbuzz-subset -lexpat -lbrotlidec -lbrotlicommon"
  CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} go build #编译
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

### 使用说明同[主线](https://github.com/MkvAutoSubset/MkvAutoSubset?tab=readme-ov-file#%E4%B8%80%E9%83%A8%E5%88%86%E4%B8%AD%E6%96%87%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E%E8%8B%B1%E6%96%87%E5%AE%8C%E6%95%B4%E7%89%88)