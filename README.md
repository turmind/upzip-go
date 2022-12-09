# S3文件上传自动解压缩

## 创建函数

- 到lambda控制台，点击创建函数
- 选择从头开始创作，输入函数名
- 选择运行环境为go1.x
- 其他默认，进行创建函数

## 修改函数常规配置

- 点击编辑，修改内存大小为4096，短暂存储为2048，超时时长为10分钟，保存
- 内存大小影响CPU的分配，4096的情况下，能够保证有2个CPU以上的分配，版本为多协程运行，可最大利用CPU
- 短暂存储为解压使用，可根据压缩包的大小选择合适的大小
- 超时时长按需要设置
- 环境变量设置，增加环境变量RATE，设置为200，环境变量用于控制并发访问速率

## 设置运行权限

- 点击权限tab，通过角色名称链接跳转到IAM页面
- 为角色添加权限：AmazonS3FullAccess 及 AWSLambdaVPCAccessExecutionRole
- AmazonS3FullAccess权限较宽，可根据需要自定义策略，这里只为演示，直接选择fullaccess
- AWSLambdaVPCAccessExecutionRole为lambda运行在vpc内需要

## 设置VPC

- 选择lambda运行时所在的vpc、子网、安全组
- 所选择的子网所关联的路由表需要确认具有到S3 endpoint gateway的路由

## 代码编译及上传

```linux
GOARCH=amd64 GOOS=linux go build main.go
zip main.zip main
```

- 将main.zip直接上传
- 代码运行时设置，编辑修改处理程序为 “main”

## 添加触发器

- 选择S3,输入bucket名称并选择
- Event type选择ALL event
- Suffix 输入 .zip
- 勾选 i acknoledge并点击添加

## 测试

- 向对应的S3 bucket上传 .zip文件，查看文件是否正确解压

## 注意事项

- 当zip包中存在zip包，将导致包中的zip包也继续解压
- 代码作为示例使用，需要根据自身需求进一步调整
