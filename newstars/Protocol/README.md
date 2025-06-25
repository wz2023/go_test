# 协议说明：
  本协议库采取Protocol Buffers 3.4.0 版本编译生成，规范请参照 Protocol Buffers 3.4.0 官方说明

# 依赖：
	Golang 1.9.0     或更高版本
	Node.js 6.11.2   或更高版本
	npm 3.10.10      或更高版本
	google-protobuf  安装命令 npm install google-protobuf
	webpack          安装命令 npm install webpack -g
	protoc 3.4.0     自带

## 协议明明规范：
	所有协议名约定由1位字母和7位16进制数字组成，例如:S3010001
	首字母约定:
		1:S   服务端发送的协议
		2:C   客户端发送的协议
		3:N   客户端通知的协议
		4:P   服务端推送的协议
		5:I   服务端或客户端内部通讯协议
		6-F   保留类型
	第二位约定:
		0     系统或架构级别的协议
		1     通用协议
		2     未使用
		3     内部游戏协议（内部游戏协议三四位表示游戏类型，例如01表示斗地主，余下各位为协议标识号）
		4-F   未使用

## 完整包格式

| Type     | Length   | Ordinal | Protocol body |
| -------- | -------- | ------- | ------------- |
| 1 byte   | 3 byte   | 1 byte  | S3010001...   |

    Type    协议类型：
        0：protobuf
    Length  协议包体长度，大端整数
    Ordinal 包序(0~255)
    Body    消息内容
