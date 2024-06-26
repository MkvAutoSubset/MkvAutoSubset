# C导出函数说明

## 2022.05新增生成测试视频说明

- ```c
  bool CreateBlankOrBurnVideo(long t, char* s, char* enc, char* ass, char* fontdir, char* output);
  //创建一个空视频或者烧录字幕的视频
  //t: 视频时长
  //s: 源视频路径(留空即生成空视频)
  //enc: 视频编码器
  //ass: 字幕文件路径(当s为空时,t参数自动设置为字幕时长)
  //fontdir: 字体目录路径
  //output: 输出文件
  //return: 是否成功完成
  ```
- ```c
  bool CreateTestVideo(char* asses, char* s, char* fontdir, char* enc, bool burn, logCallback lcb);
  //创建测试视频
  //asses: 字幕文件数组的json格式文本
  //s: 源视频路径
  //fontdir: 字体目录路径
  //enc: 视频编码器
  //burn: 是否烧录字幕
  //return: 是否成功完成
  ```

## 2022.04新增检查模式说明

- ```c
  void Check(bool check, bool strict);
  //启用检查模式(影响包含子集化操作的工作流)
  //check: 是否启用检查模式
  //strict: 是否启用严格模式
  ```

## 2022.04更新的ASS转PGS说明

- ```c
  void A2P(bool a2p, bool apc, char* pr, char* pf);
  //启用ass转pgs(依赖ass2bdnxml且应在执行工作流之前调用)
  //a2p: 是否启用ass转pgs
  //apc: 是否使ass与pgs共存
  //pr: pgs分辨率
  //pf: pgs帧率
  ```

## 日志回调

- 原型
    ```c
    void (*logCallback)(unsigned char l, char* str); 
    //l: 日志等级(0:Info, 1:Warning, 2:SWarning, 3:Error, 4:Progress)
    //str: UTF-8编码的指针,并约定所有"char*"数据类型的参数或返回值都为此.
    ```
- 一些说明
    - 几乎所有导出的方法都有这个参数(在最后),当出现错误时会进行调用,可以用来判断执行过程是否出错,错在哪.
    - 虽然可以为NULL,但并不建议这样做.
    - 以下名为"lcb"的参数均为日志回调,不再赘述.
## 版本信息
- ```c
  char* Version();
  //return: 库版本信息
  ```
## 初始化实例

- ```c
  bool InitInstance(logCallbac lcb);
  //return: 是否初始化成功
  ```
- 应该被最先调用.
- 会检测依赖,如果不满足会返回false.
- 如果在**未**或**未成功**调用本函数的情况下调用其他函数(Version除外)会永远返回失败状态.

### 缓存相关

- ```c
    void Cache(char* ccs);
    //设置字体缓存(应在执行工作流之前调用)
    //p: 包含缓存文件路径的json化文本
    ```
- ```c
    char* CreateFontsCache(char* dir, char* output, logCallback lcb);
    //从字体目录创建缓存
    //dir: 字体文件目录
    //output: 缓存文件保存路径
    //return: 缓存失败字体的json格式的数组
    ```
- ```c
    bool CopyFontsFromCache(char* asses, char* dist, logCallback lcb);
    //从缓存复制字幕所需的字体
    //asses: 字幕文件路径的json的数组
    //dist: 字体文件保存目录
    //return: 是否全部导出
    ```

### 查询相关

- ```c
  char* GetFontInfo(char* p);
  //查询一个字体的信息
  //p: 字体文件路径
  //return: json格式的文件信息,如果出错会返回"null".
  ```
- ```c
  char* GetMKVInfo(char* file);
  //查询一个mkv文件内封的字幕和字体信息
  //file: 文件路径
  //return: json格式的字体信息,如果出错会返回"null".
  ```
- ```c
  char* CheckSubset(char* file, logCallbac lcb);
  //查询一个mkv文件是否需要子集化操作
  //file: mkv文件路径
  //return: 包含是否已子集化和是否出错两个bool成员的json文本.
  ```
- ```c
  char* QueryFolder(char* dir, logCallbac lcb);
  //查询一个文件夹里的mkv文件是否需要子集化
  //dir: 文件夹路径
  //return: 需要子集化的mkv文件路径数组
  ```

### MKV相关

- ```c
  void MKS(bool mks);
  //使用mks输出
  //mks: 是否启用
  ```
- ```c
  bool DumpMKV(char* file, char* output, bool subset, logCallback lcb);
  //抽取一个mkv文件里的字幕和字体并顺便进行子集化(可选)
  //file: mkv文件路径
  //output: 输出文件夹路径
  //subset: 是否进行子集化
  //return: 是否全程无错
  ```
- ```c
  bool DumpMKVs(char* dir, char* output, bool subset, logCallback lcb);
  //抽取一个文件夹里的mkv的字幕和字体并顺便进行子集化(可选)
  //dir: 文件夹路径
  //output: 输出文件夹路径
  //subset: 是否进行子集化
  //return: 是否全程无错
  ```
    - 输出文件夹的目录结构请参考[这里](https://github.com/MkvAutoSubset/MkvAutoSubset#mkvtool-%E5%8A%9F%E8%83%BD%E5%8F%8A%E4%BD%BF%E7%94%A8%E7%A4%BA%E4%BE%8B)
- ```c
  bool CreateMKV(char* file, char* tracks, char* attachments, char* output, char* slang, char* stitle, bool clean);
  //将字幕和字体封进mkv文件
  //file: 源文件路径(并非一定要是mkv文件,其他视频文件也可.)
  //tracks: 字幕文件路径数组的json化文本
  //attachments: 字体文件路径数组的json化文本
  //output: 输出文件路径
  //slang: 默认字幕语言
  //stitle: 默认字幕标题
  //clean: 是否清除源mkv原有的字幕和字体
  //return: 是否全程无错
  ```
    - 关于字幕的命名方式请参考[这里](https://github.com/MkvAutoSubset/MkvAutoSubset#%E4%B8%80%E4%BA%9B%E7%A2%8E%E7%A2%8E%E5%BF%B5)
- ```c
  bool CreateMKVs(char* vDir, char* sDir, char* fDir, char* tDir, char* oDir, char* slang, char* stitle, bool clean, logCallback lcb);
  //从一组文件夹获得情报自动生成一组mkv并自动进行子集化操作
  //vDir: 视频文件夹路径
  //sDir: 字幕文件夹路径
  //fDir: 字体文件夹路径
  //tDir: 子集化数据临时存放文件夹路径(如果为空字符串则自动指定到系统临时文件夹如"/tmp")
  //oDir: 成品输出文件夹路径
  //slang: 默认字幕语言
  //stitle: 默认字幕标题
  //clean: 是否清除源mkv原有的字幕和字体
  //return: 是否全程无错
  ```
    - 关于字幕的命名方式请参考[这里](https://github.com/MkvAutoSubset/MkvAutoSubset#%E4%B8%80%E4%BA%9B%E7%A2%8E%E7%A2%8E%E5%BF%B5)
- ```c
  bool MakeMKVs(char* dir, char* data, char* output, char* slang, char* stitle, bool subset, logCallback lcb);
  //用子集化后的数据目录替代原有的字幕和字体
  //dir: 源mkv集合文件夹路径
  //data: 子集化后的数据文件夹路径
  //output: 新mkv集合输出文件夹路径
  //slang: 默认字幕语言
  //stitle: 默认字幕标题
  //subset: 是否进行子集化
  //return: 是否全程无错
  ```
    - 输出文件夹的目录结构请参考[这里](https://github.com/MkvAutoSubset/MkvAutoSubset#mkvtool-%E5%8A%9F%E8%83%BD%E5%8F%8A%E4%BD%BF%E7%94%A8%E7%A4%BA%E4%BE%8B)
    - 关于字幕的命名方式请参考[这里](https://github.com/MkvAutoSubset/MkvAutoSubset#%E4%B8%80%E4%BA%9B%E7%A2%8E%E7%A2%8E%E5%BF%B5)

### 字幕相关

- ```c
    bool ASSFontSubset(char* files, char* fonts, char* output, bool dirSafe, logCallback lcb);
    //对字幕和字体进行子集化操作
    //files: 字幕文件路径数组的json化文本
    //fonts: 字体文件夹路径
    //output: 成品输出文件夹路径
    //dirSafe: 是否把成品输出到"${output}/subsetted"文件夹里(为了安全建议设置为true)
    //return: 是否全程无错
  ```
- ```c
    char* GetFontsList(char* files, char* fonts, logCallback lcb);
    //取得数组内字幕需要的全部字体,如果设置了Check则会试图匹配字体,并输出匹配失败的列表.
    //files: 字幕文件路径的json的数组
    //fonts: 字体文件夹路径
    //return: json格式的二维数组(第一个成员是需要的字体名称,第二个成员是没有匹配成功的字体名称.)
  ```
- ```c
  void NRename(bool n);
  //子集化时不重命名字体
  //n: 是否不重命名
   ```
- ```c
  void NOverwrite(bool n);
  //输出时是否跳过已存在的文件
  //o: 是否跳过
   ```