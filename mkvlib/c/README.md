# C导出函数说明

## 日志回调
- 原型
    ```c
    void (*logCallbac)k(char* str); //str: UTF-8编码的指针,并确定所有"char*"数据类型的参数或返回值都为此.
    ```
- 一些说明
  - 几乎所有导出的方法都有这个参数(在最后),当出现错误时会进行调用,可以用来判断执行过程是否出错,错在哪.
  - 虽然可以为NULL,但并不建议这样做.
  - 以下名为"lcb"的参数均为日志回调,不再赘述.

## 初始化实例
- *InitInstance(logCallbac lcb)*
- 应该被最先调用.
- 会检测依赖,如果不满足会返回false.
- 如果在未调用本函数的情况下调用其他函数会永远返回失败状态.

### 查询相关
- ```c
  char* GetMKVInfo(char* file);
  //查询一个mkv文件内封的字幕和字体信息
  //file: 文件路径
  //return: json格式的文件信息,如果出错会返回"null".
  ```
- ```c
  char* CheckSubset(char* file,logCallbac lcb);
  //查询一个mkv文件是否需要子集化操作
  //file: mkv文件路径
  //return: 包含是否已子集化和是否出错两个bool成员的json文本.
  ```
- ```c
  char* QueryFolder(char* dir,logCallbac lcb);
  //查询一个文件夹里的mkv文件是否需要子集化
  //dir: 文件夹路径
  //return: 需要子集化的mkv文件路径数组
  ```