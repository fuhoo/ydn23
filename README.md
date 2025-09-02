# YDN23 Protocol Implementation

YD/T 1363 通信局(站)电源、空调及环境集中监控管理系统 第3部分：前端智能设备协议。
YDN23 电总协议实现，支持协议帧的编码和解码功能。

## 功能特性

- ✅ 协议帧编码 (Encode)
- ✅ 协议帧解码 (Decode) 
- ✅ CHKSUM 校验和计算
- ✅ LENGTH 字段验证
- ✅ 完整的单元测试

## 安装

```bash
go get github.com/fuhoo/ydn23
```

## 使用方法

### 编码协议帧 (Encode)

```go
package main

import (
    "fmt"
    "github.com/fuhoo/ydn23"
)

func main() {
    // 创建协议数据单元
    pdu := &ydn23.ProtocolDataUnit{
        Ver:  0x21,  // 协议版本
        Addr: 0x01,  // 设备地址
        Cid1: 0x60,  // 控制标识码1
        Cid2: 0x42,  // 控制标识码2 (获取模拟量数据)
        Info: []byte{}, // 信息数据
    }
    
    // 编码为协议帧
    frame, err := ydn23.Encode(pdu)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("编码结果: % X\n", frame)
    // 输出: 7E 32 31 30 31 36 30 34 32 30 30 30 30 30 46 44 42 32 0D
}
```

### 解码协议帧 (Decode)

```go
package main

import (
    "fmt"
    "github.com/fuhoo/ydn23"
)

func main() {
    // 协议帧数据 (获取模拟量数据命令)
    frame := []byte{0x7E, 0x32, 0x31, 0x30, 0x31, 0x36, 0x30, 0x34, 0x32, 0x30, 0x30, 0x30, 0x30, 0x30, 0x46, 0x44, 0x42, 0x32, 0x0D}
    
    // 解码协议帧
    pdu, err := ydn23.Decode(frame)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("协议版本: 0x%02X\n", pdu.Ver)
    fmt.Printf("设备地址: 0x%02X\n", pdu.Addr)
    fmt.Printf("控制标识码1: 0x%02X\n", pdu.Cid1)
    fmt.Printf("控制标识码2: 0x%02X\n", pdu.Cid2)
    fmt.Printf("信息数据: % X\n", pdu.Info)
}
```

### 带数据的响应帧示例

```go
// 编码带INFO数据的响应帧
responsePDU := &ydn23.ProtocolDataUnit{
    Ver:  0x21,
    Addr: 0x01,
    Cid1: 0x60,
    Cid2: 0x00, // 正常响应
    Info: []byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35}, // 模拟量数据
}

responseFrame, err := ydn23.Encode(responsePDU)
if err != nil {
    panic(err)
}

fmt.Printf("响应帧: % X\n", responseFrame)
```

## 协议格式

协议帧基本格式：

| 字段 | 字节数 | 说明 |
|------|--------|------|
| SOI | 1 | 起始位标志 (0x7E) |
| VER | 1 | 协议版本号 (0x21) |
| ADR | 1 | 设备地址 |
| CID1 | 1 | 控制标识码1 |
| CID2 | 1 | 控制标识码2/返回码 |
| LENGTH | 2 | INFO字节长度 |
| INFO | LENID/2 | 信息数据 |
| CHKSUM | 2 | 校验和码 |
| EOI | 1 | 结束码 (0x0D) |

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0x00 | 正常 |
| 0x01 | VER值错误 |
| 0x02 | CHKSUM错误 |
| 0x03 | LCHKSUM错误 |
| 0x04 | CID2无效 |
| 0x05 | 命令格式错误 |
| 0x06 | 数据无效 |
| 0x80~EF | 其它错误 |

## 测试

运行单元测试：

```bash
go test
```

运行特定测试：

```bash
go test -v -run TestDecode
go test -v -run TestEncode
```

## 许可证

MIT License
