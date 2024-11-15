# ias-kit

## 功能列表
- oss
- zip_commenter   
为 zip 文件添加 comment 信息   
安装：   
go install gitlab.ctyuncdn.cn/ias/ias-kit/zip_commenter    
添加 comment 信息:   
zip_commenter -i foo.zip -c meta.json -m w   
读取 zip comment 信息:   
zip_commenter -i foo.zip -m r