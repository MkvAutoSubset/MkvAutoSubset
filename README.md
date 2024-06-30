# Mkv Auto Subset

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/MkvAutoSubset/MkvAutoSubset?include_prereleases)

ASS字幕字体子集化 MKV批量提取/生成

[问题反馈](https://bbs.acgrip.com/thread-9070-1-1.html)

[next版本](https://github.com/MkvAutoSubset/MkvAutoSubset/tree/next):无需安装FontTools,自带ASS转PGS.

## 什么叫字幕字体子集化
- 这里说的字幕特指ass(ssa)这种带有特效的文本字幕;
- ass字幕会引用一些字体,这些字体在播放器所在的系统里可能有安装,也可能没有;
- 为了实现在任意地方都能有完整的视觉体验,可以把字幕以及字幕里所引用的字体文件一起打包进mkv文件里;
- 以上的操作存在一个问题,有些字幕会引用很多字体,这些字体文件体积动辄几十MB,而字幕只用到了其中的几个字而已;
- 比如一个番剧本体200M,但打包了字体文件后变成400M了,这像画吗?
- 综上所述,子集化的目的就是把字体拆包,找出字幕用到的那部分字形并重新打包;
- 好处不仅限于节约存储空间,加快缓冲速度;
- 想想看:在一个只有30Mbps上传的网络环境下,要看上面那个光字体就200M的番剧,这河里吗?

## mkvtool 安装

### Win64依赖打包说明
- 点[这里](https://github.com/MkvAutoSubset/MkvAutoSubset/releases/download/win64_assets/win64_assets.zip)下载
- 解压到本地
- 运行
  ```shell 
  intall.bat
  ```

### Docker镜像使用说明
- 从Dockerhub获取
  ```shell
  TAGNAME=master
  FONT_DIR="/usr/share/fonts/truetype" #字体目录
  CACHE_DIR="${HOME}/.mkvtool/caches"  #缓存目录
  OTHER_DIR="" #其他目录(可选,示例见下节.)
  docker pull ac79b0c6/mkvtool:${TAGNAME} #拉取/更新镜像
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

### 依赖

- FontTools
  ```shell
  apt install fonttools #Debian/Ubuntu
  apk add py3-fonttools #Alpine
  brew install fonttools #macOS
  pip install fonttools #Use pip
  ```
- MKVToolNix
  ```shell
  apt install mkvtoolnix #Debian/Ubuntu
  apk add mkvtoolnix #Alpine
  brew install mkvtoolnix #macOS
  ```
- ass2bdnxml

  从[这里](https://github.com/Masaiki/ass2bdnxml/releases)获取

#### 关于Windows用户

- 从 [这里](https://www.python.org/downloads) 下载并安装Python
- 命令提示符(CMD)里参考上面使用pip的方式安装FontTools依赖
- 从 [这里](https://github.com/Masaiki/ass2bdnxml/releases) 获取ass2bdnxml
- 从 [这里](https://www.fosshub.com/MKVToolNix.html) 下载并安装MKVToolNix
- 保证以上两个依赖项的相关可执行文件(_ttx.exe_,_pyftsubset.exe_,_mkvextract.exe_,_mkvmerge.exe_,_ass2bdnxml.exe_)在 **path** 环境变量里

### 本体

- 有安装Go的情况:
  ```shell
  go install github.com/MkvAutoSubset/MkvAutoSubset/mkvtool@latest #安装和更新
  ```

- Arch Linux用户(通过Arch User Repository):
  - 点击[这里](https://aur.archlinux.org/packages/mkvtool/) 查看具体信息或使用AUR Helper
  ```shell
  yay -S mkvtool #yay
  paru -S mkvtool #paru
  ```

- 手动安装:

  [点此下载](https://github.com/MkvAutoSubset/MkvAutoSubset/releases/latest)
- 适用于Win64的GUI版及动态链接库

  [点此下载](https://github.com/MkvAutoSubset/MkvAutoSubset/releases/gui)
### 一部分中文使用说明([英文完整版](./mkvtool/docs/mkvtool.md))
- 旧版CLI中"标准工作流"的替代
  ```shell
  mkvtool d bangumi && mkvtool m bangumi #假设mkv文件在"bangumi"文件夹中
  ```
- 对单(或多)个(或文件夹内的)字幕进行子集化
  ```shell
  mkvtool s test.ass #单个文件
  mkvtool s 01.ass 02.ass #多个文件
  mkvtool s subs #文件夹
  ```
- 查看某个字体的信息
  ```shell
  mkvtool i font.ttf
  ```
- 检测单个字体文件(或目录)需要哪些字体
  ```shell
  mkvtool l test.ass #单个文件
  mkvtool l subs #目录
  ```
- 从单个(或文件夹的)mkv文件里抽取字幕和字体*并创建子集化后的版本(可选)*
  ```shell
  mkvtool d file.mkv #单个文件
  mkvtool d bangumi #文件夹
  
  #可选"-n"参数:当"-n"存在时,只抽取内容,不进行子集化操作.
  ```
- 检测单个(或文件夹的)mkv文件字幕和字体,判断是否需要子集化
  ```shell
  mkvtool q file.mkv #单个文件,会直接输出是否需要子集化
  mkvtool q bangumi #文件夹,会将需要子集化的文件列表输出至"${workdir}/result.txt".
  ```
- 将子集化后的字幕与字体替代原有的内容
  ```shell
  mkvtool m bangumi dist

  #假设bangumi文件夹里的目录结构如下所示:
  #bangumi
  # |-- S01
  # ||-- abc S01E01.mkv
  # ||-- abc SxxExx.mkv
  # |-- SP.mkv
  # |-- xx.mkv
  #那么对应的data文件夹的目录结构应该是如下的所示:
  #data
  # |-- S01
  # ||-- abc S01E01
  # |||-- ...
  # |||-- subsetted
  # |||-- xxx.sub
  # ||-- abc SxxExx
  # |||-- ...
  # |||-- subsetted
  # |||-- xxx.sub
  # |||-- ...
  # |-- SP
  # |||-- ...
  # |||-- subsetted
  # |||-- xxx.sub
  # |||-- ...
  # |-- xx
  # |||-- ...
  # |||-- subsetted
  # |||-- xxx.sub
  # |||-- ...
  
  #*奇淫巧技:指定一个没有任何内容的data文件夹,将输出一个"干净"的mkv文件.
   ```
- 从一组文件夹获得情报并生成一组mkv
  ```shell
  mkvtool c bangumi
  
  #可选"-c"参数:当"-c"存在时,将清空原有的字幕和字体(默认为追加).
  #bangumi文件夹里的目录结构应如下所示:
  #bangumi
  # |-- v
  # ||-- aaa.mkv
  # ||-- bbb.mp4
  # ||-- ccc.avi
  # |-- s
  # ||-- aaa.ass
  # ||-- aaa.srt
  # ||-- aaa.sup
  # ||-- aaa.xxx
  # ||-- bbb.xxx
  # ||-- ccc.xxx
  # |-- f
  # ||-- abc.ttf
  # ||-- def.ttc
  # ||-- ghi.otf
  # ||-- ...
  
  #若遇到ass字幕会自动进行子集化操作.
  #成品会放在"${bangumi}/o"文件夹中.
  ```
  
### 一些碎碎念
- 手动指定缓存文件夹路径,当提供的字体目录里缺少字体时,会尝试在缓存里查找.
- 输出终端输出到指定文件,空为不输出,默认为空.
- 字幕文件名规范:
  ```
  抽取出来的字幕长得像是如下的样子:
  #a_b_c.d
  #:如果文件名以"#"开头,代表这个轨道是默认轨道.
  a:轨道编号(在"c"模式里,这里应该和视频文件的文件名相同.)
  b:字幕语言代码
  c:字幕标题
  d:字幕文件后缀名
  
  那么,请体会在"c"模式中,以下的命名方式所带来的便利:
  |-- v
  ||-- aaa.mp4
  |-- s
  ||-- #aaa_chi_简体中文.ass
  ||-- aaa_chi_繁體中文.srt
  ||-- aaa_jpn_日本語.sup
  ||-- aaa_eng_English.srt
  ```
- 字幕语言代码表:

  [点此获取](https://www.science.co.il/language/Codes.php)


## 警告
**不要使用特殊字符和引号，以避免字符串分割和子文件夹问题**

轨道名称中包含 `/` 或其他特殊字符会导致 mkvtool 出错.同样,字体名称中包含 `'!#` 或其他特殊字符也会有问题.命令行中的参数不会为 mkvmerge 进行引用和转义.