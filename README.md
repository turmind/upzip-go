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

## 成功、失败通知

- 在lambda中通过添加目标的方式，将结果转发到sns的topic
- 选择时，源异步调用即可，条件根据需要可以选择成功时或者失败时或两者的情况下都发送
- 目标类型选择SNS主题，通过SNS可以将消息发送到email或者lambda,lambda通过代码编写的方式可将消息转发到：slack/wechat/dingding/lark
- 配置SNS邮件通知可参考链接<https://docs.aws.amazon.com/zh_cn/sns/latest/dg/sns-email-notifications.html>
- 以下为消息参考：

成功

```json
{
	"version": "1.0",
	"timestamp": "2022-12-09T09:45:19.739Z",
	"requestContext": {
		"requestId": "3cea2328-f72b-48ea-b0c1-d02fc73d90d7",
		"functionArn": "arn:aws:lambda:ap-southeast-1:900212707297:function:unzip-golang:$LATEST",
		"condition": "Success",
		"approximateInvokeCount": 1
	},
	"requestPayload": {
		"Records": [{
			"eventVersion": "2.1",
			"eventSource": "aws:s3",
			"awsRegion": "ap-southeast-1",
			"eventTime": "2022-12-09T09:45:18.417Z",
			"eventName": "ObjectCreated:Put",
			"userIdentity": {
				"principalId": "AWS:AIDA5DGG3Z7QQ3UIU4GTS"
			},
			"requestParameters": {
				"sourceIPAddress": "54.240.199.105"
			},
			"responseElements": {
				"x-amz-request-id": "QYJBYXS10RMAP0DD",
				"x-amz-id-2": "bMPEyOa5v6/+qys20CjapRvwqhG1ECWui19YMRA4wKZpm7kTmRpeDR7n8b/RoTVymspselGiZby+a08dwDikHhEn9Dc52UOe"
			},
			"s3": {
				"s3SchemaVersion": "1.0",
				"configurationId": "b42ccd47-2338-43ca-af3c-0e4c5f8452db",
				"bucket": {
					"name": "unzip-file",
					"ownerIdentity": {
						"principalId": "A1TE1NMOB9GT94"
					},
					"arn": "arn:aws:s3:::unzip-file"
				},
				"object": {
					"key": "abc/dd/apk.zip",
					"size": 25934,
					"eTag": "9db5d3f2b5716133719e0a8b4ac44c8b",
					"sequencer": "00639303AE5B5E97DB"
				}
			}
		}]
	},
	"responseContext": {
		"statusCode": 200,
		"executedVersion": "$LATEST"
	},
	"responsePayload": "[aws:s3 - 2022-12-09 09:45:18.417 +0000 UTC] Bucket = unzip-file, Key = abc/dd/apk.zip 
	2022 - 12 - 09 09: 45: 19.54130159 + 0000 UTC m = +259.205749786 download zip sucess
	2022 - 12 - 09 09: 45: 19.615593778 + 0000 UTC m = +259.280041992 total file upload: 3,
	success: 3 "}
```

失败

```json
{
	"version": "1.0",
	"timestamp": "2022-12-09T09:46:39.136Z",
	"requestContext": {
		"requestId": "4b4ab4f3-0d7b-4636-ba92-68addf470346",
		"functionArn": "arn:aws:lambda:ap-southeast-1:900212707297:function:unzip-golang:$LATEST",
		"condition": "RetriesExhausted",
		"approximateInvokeCount": 3
	},
	"requestPayload": {
		"Records": [{
			"eventVersion": "2.1",
			"eventSource": "aws:s3",
			"awsRegion": "ap-southeast-1",
			"eventTime": "2022-12-09T09:43:46.099Z",
			"eventName": "ObjectCreated:Put",
			"userIdentity": {
				"principalId": "AWS:AIDA5DGG3Z7QQ3UIU4GTS"
			},
			"requestParameters": {
				"sourceIPAddress": "54.240.199.105"
			},
			"responseElements": {
				"x-amz-request-id": "MY808DYSZFNFEK88",
				"x-amz-id-2": "Hocg25k+Csi+F7RM0PcRC/TDxK+F1KLGoZyWu+Oh3D8LMaG6opLyUbElDKA2RRl2ldeMcaPCcp+6eTzePKQAWeSmN48cKxAy"
			},
			"s3": {
				"s3SchemaVersion": "1.0",
				"configurationId": "b42ccd47-2338-43ca-af3c-0e4c5f8452db",
				"bucket": {
					"name": "unzip-file",
					"ownerIdentity": {
						"principalId": "A1TE1NMOB9GT94"
					},
					"arn": "arn:aws:s3:::unzip-file"
				},
				"object": {
					"key": "abc/cc.zip",
					"size": 224448,
					"eTag": "bafcaedda0837c5da26db29c04015b20",
					"sequencer": "00639303520601BA44"
				}
			}
		}]
	},
	"responseContext": {
		"statusCode": 200,
		"executedVersion": "$LATEST",
		"functionError": "Unhandled"
	},
	"responsePayload": {
		"errorMessage": "zip: not a valid zip file",
		"errorType": "errorString"
	}
}
```
