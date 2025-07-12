package ydn23

import (
	"fmt"
)

// CalcCHKSUM 计算 CHKSUM 校验和
// data: 不包含 SOI（如'~'）、EOI（如"CR"）、CHKSUM（如"FD3B"）的字符串
func CalcCHKSUM(data string) string {
	var sum uint16 = 0
	for i := 0; i < len(data); i++ {
		sum += uint16(data[i])
	}
	// 取反加1（补码）
	chksum := ^sum + 1
	// 只保留16位
	chksum &= 0xFFFF
	// 返回4位大写十六进制字符串
	return fmt.Sprintf("%04X", chksum)
}
