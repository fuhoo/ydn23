package ydn23

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	// 测试用例1：获取模拟量数据命令帧
	// ~210160420000FDB2\r
	testFrame1 := []byte{0x7E, 0x32, 0x31, 0x30, 0x31, 0x36, 0x30, 0x34, 0x32, 0x30, 0x30, 0x30, 0x30, 0x30, 0x46, 0x44, 0x42, 0x32, 0x0D}

	pdu, err := Decode(testFrame1)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// 验证解码结果
	if pdu.Ver != 0x21 {
		t.Errorf("Expected Ver=0x21, got 0x%02X", pdu.Ver)
	}
	if pdu.Addr != 0x01 {
		t.Errorf("Expected Addr=0x01, got 0x%02X", pdu.Addr)
	}
	if pdu.Cid1 != 0x60 {
		t.Errorf("Expected Cid1=0x60, got 0x%02X", pdu.Cid1)
	}
	if pdu.Cid2 != 0x42 {
		t.Errorf("Expected Cid2=0x42, got 0x%02X", pdu.Cid2)
	}
	if len(pdu.Info) != 0 {
		t.Errorf("Expected empty Info, got length %d", len(pdu.Info))
	}

	// 测试用例2：带INFO数据的响应帧
	// 模拟响应帧：~2101600000000C30313233343536373839303132FDB2\r
	// VER=21, ADR=01, CID1=60, CID2=00, LENGTH=000C, INFO=30313233343536373839303132
	testFrame2 := []byte{0x7E, 0x32, 0x31, 0x30, 0x31, 0x36, 0x30, 0x30, 0x30, 0x30, 0x30, 0x43, 0x33, 0x30, 0x33, 0x31, 0x33, 0x32, 0x33, 0x33, 0x33, 0x34, 0x33, 0x35, 0x33, 0x36, 0x33, 0x37, 0x33, 0x38, 0x33, 0x39, 0x33, 0x30, 0x33, 0x31, 0x33, 0x32, 0x46, 0x44, 0x42, 0x32, 0x0D}

	pdu2, err := Decode(testFrame2)
	if err != nil {
		t.Fatalf("Decode testFrame2 failed: %v", err)
	}

	if pdu2.Ver != 0x21 {
		t.Errorf("Expected Ver=0x21, got 0x%02X", pdu2.Ver)
	}
	if pdu2.Addr != 0x01 {
		t.Errorf("Expected Addr=0x01, got 0x%02X", pdu2.Addr)
	}
	if pdu2.Cid1 != 0x60 {
		t.Errorf("Expected Cid1=0x60, got 0x%02X", pdu2.Cid1)
	}
	if pdu2.Cid2 != 0x00 {
		t.Errorf("Expected Cid2=0x00, got 0x%02X", pdu2.Cid2)
	}
	expectedInfo := []byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32}
	if !bytes.Equal(pdu2.Info, expectedInfo) {
		t.Errorf("Expected Info=%v, got %v", expectedInfo, pdu2.Info)
	}
}

func TestEncode(t *testing.T) {
	// 测试用例1：编码获取模拟量数据命令
	pdu1 := &ProtocolDataUnit{
		Ver:  0x21,
		Addr: 0x01,
		Cid1: 0x60,
		Cid2: 0x42,
		Info: []byte{},
	}

	encoded1, err := Encode(pdu1)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	fmt.Printf("encoded1: % X, ASCII: %s\n", encoded1, encoded1)
	// 验证编码结果
	expected1 := []byte{0x7e, 0x32, 0x31, 0x30, 0x31, 0x36, 0x30, 0x34, 0x32, 0x30, 0x30, 0x30, 0x30, 0x46, 0x44, 0x42, 0x30, 0x0d}

	if !bytes.Equal(encoded1, expected1) {
		t.Errorf("Expected encoded frame % X, got % X", expected1, encoded1)
	}

	// 测试用例2：编码带INFO数据的响应
	pdu2 := &ProtocolDataUnit{
		Ver:  0x21,
		Addr: 0x01,
		Cid1: 0x60,
		Cid2: 0x00,
		Info: []byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32},
	}

	encoded2, err := Encode(pdu2)
	if err != nil {
		t.Fatalf("Encode pdu2 failed: %v", err)
	}
	fmt.Printf("encoded2:% X,ASCII: %s\n", encoded2, encoded2)

	// 验证编码结果（注意CHKSUM会不同，因为INFO内容不同）
	if len(encoded2) < 9 {
		t.Errorf("Encoded frame too short: %d", len(encoded2))
	}
	if encoded2[0] != 0x7E {
		t.Errorf("Expected SOI=0x7E, got 0x%02X", encoded2[0])
	}
	if encoded2[len(encoded2)-1] != 0x0D {
		t.Errorf("Expected EOI=0x0D, got 0x%02X", encoded2[len(encoded2)-1])
	}

	
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	// 测试编码解码往返
	originalPDU := &ProtocolDataUnit{
		Ver:  0x21,
		Addr: 0x01,
		Cid1: 0x60,
		Cid2: 0x42,
		Info: []byte{0x01, 0x02, 0x03, 0x04},
	}

	// 编码
	encoded, err := Encode(originalPDU)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// 解码
	decodedPDU, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// 验证往返一致性
	if originalPDU.Ver != decodedPDU.Ver {
		t.Errorf("Ver mismatch: original=0x%02X, decoded=0x%02X", originalPDU.Ver, decodedPDU.Ver)
	}
	if originalPDU.Addr != decodedPDU.Addr {
		t.Errorf("Addr mismatch: original=0x%02X, decoded=0x%02X", originalPDU.Addr, decodedPDU.Addr)
	}
	if originalPDU.Cid1 != decodedPDU.Cid1 {
		t.Errorf("Cid1 mismatch: original=0x%02X, decoded=0x%02X", originalPDU.Cid1, decodedPDU.Cid1)
	}
	if originalPDU.Cid2 != decodedPDU.Cid2 {
		t.Errorf("Cid2 mismatch: original=0x%02X, decoded=0x%02X", originalPDU.Cid2, decodedPDU.Cid2)
	}
	if !bytes.Equal(originalPDU.Info, decodedPDU.Info) {
		t.Errorf("Info mismatch: original=%v, decoded=%v", originalPDU.Info, decodedPDU.Info)
	}
}
