# Mkv Auto Subset

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/KurenaiRyu/MkvAutoSubset?include_prereleases)

自动字体子集化工具

## mkvtool 安装
### 依赖
- fonttools
  ```shell
  pip install fonttools
  ```
- mkvtoolnix
  ```shell
  apt install mkvtoolnix #Debian/Ubuntu
  apk add mkvtoolnix #Alpine
  ```
### 本体
- 有安装Go的情况:
  ```shell
  go install https://github.com/KurenaiRyu/MkvAutoSubset/mkvtool@latest #安装和更新
  ```
- 手动安装:

  [点此下载](https://github.com/KurenaiRyu/MkvAutoSubset/releases/latest)

## mkvtool 功能及使用示例

- 从单个(或文件夹的)mkv文件里抽取字幕和字体*并创建子集化后的版本(可选)*
  ```shell
  mkvtool -d -f file.mkv #单个文件
  mkvtool -d -s bangumi #文件夹
  #可选"-n"参数:当"-n"存在时,只抽取内容,不进行子集化操作.
  #可选"-data"参数,指定输出目录,默认输出到"${workdir}/data".
  ```
- 检测单个(或文件夹的)mkv文件字幕和字体,判断是否需要子集化.
  ```shell
  mkvtool -q -f file.mkv #单个文件,会直接输出是否需要子集化
  mkvtool -q -s bangumi #文件夹,会将需要子集化的文件列表输出至"${workdir}/list.txt".
  ```
- 将子集化后的字幕与字体替代原有的内容
  ```shell
  mkvtool -m -s bangumi -data data -dist dist
  #-data参数默认值为"${workdir}/data"
  #-dist参数默认值为"${workdir}/dist"
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
  
  #*奇淫巧技:指定一个没有任何内容的data目录,将输出一个"干净"的mkv文件.
   ```
- 从一组文件夹获得情报并生成一组mkv
  ```shell
  mkvtool -c -s bangumi
  #可选"-clean"参数:当"-clean"存在时,将清空原有的字幕和字体(默认为追加).
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
  
  #诺遇到ass字幕会自动进行子集化操作.
  #成品会放在"${bangumi}/o"文件夹中.
  ```
- 字体子集化
  ```shell
  mkvtool -a aaa.ass -bbb.ass -af fonts -ao output
  #"-a"参数可复用
  #"-af"参数为字体文件夹路径,默认值为"fonts".
  #"-ao"参数为子集化成品输出路径
  #在默认情况下,子集化后的成品输出于"${output}/subsetted"文件夹;
  #使用参数"-ans"可使其直接输出于"${output}",但会预先清空该文件夹,慎用.
  ```