package ydn23

import (
	"errors"
)

const (
	SOI byte = 0x7E
	EOI byte = 0x0D
)

// Rtn错误码常量定义
const (
	//err code
	RtnOk             byte = 0x00 // 正常
	RtnVerError       byte = 0x01 // VER值错误
	RtnChkSumError    byte = 0x02 // CHKSUM错误
	RtnLChkSumError   byte = 0x03 // LCHKSUM错误
	RtnCid2Invalid    byte = 0x04 // CID2无效
	RtnCMDFormatError byte = 0x05 // 命令格式错误
	RtnDataInvalid    byte = 0x06 // 数据无效
	RtnASCIIError     byte = 0x07 // ASC码错误

	//errors
	ErrVerError       ydn23Err = "ver error"
	ErrChkSumError    ydn23Err = "chksum error"
	ErrLChkSumError   ydn23Err = "lchksum error"
	ErrCid2Invalid    ydn23Err = "cid2 invalid"
	ErrCMDFormatError ydn23Err = "cmd format error"
	ErrDataInvalid    ydn23Err = "data invalid"
	ErrASCIIError     ydn23Err = "ascii error"
	ErrOtherError     ydn23Err = "other error, user-defined error"
)

const ()

type ydn23Err string

func (err ydn23Err) Error() string {
	return string(err)
}

// RTN错误码对应的错误信息
func mapRTNCodeToError(RTNCode uint8) (err error) {
	switch RTNCode {
	case RtnOk:
		err = nil
	case RtnVerError:
		err = ErrVerError
	case RtnChkSumError:
		err = ErrChkSumError
	case RtnLChkSumError:
		err = ErrLChkSumError
	case RtnCid2Invalid:
		err = ErrCid2Invalid
	case RtnCMDFormatError:
		err = ErrCMDFormatError
	case RtnDataInvalid:
		err = ErrDataInvalid
	case RtnASCIIError:
		err = ErrASCIIError
	default:
		//其它错误，用户自定义错误
		if RTNCode >= 0x80 && RTNCode <= 0xEF {
			return ErrOtherError
		}
	}
	return
}

type ProtocolDataUnit struct {
	Ver  byte
	Addr byte
	Cid1 byte
	Cid2 byte // CID2/RTN
	Info []byte
}

// 单字节转ASCII码（高4位、低4位分别转为ASCII字符）
func byteToASCIIHex(b byte) (byte, byte) {
	hi := (b >> 4) & 0x0F
	lo := b & 0x0F
	return "0123456789ABCDEF"[hi], "0123456789ABCDEF"[lo]
}

// 多字节转ASCII码
func bytesToASCIIHex(data []byte) []byte {
	res := make([]byte, 0, len(data)*2)
	for _, b := range data {
		hi, lo := byteToASCIIHex(b)
		res = append(res, hi, lo)
	}
	return res
}

// LENGTH 字段计算
func calcLENGTH(lenid uint16) (asciiLen []byte) {
	// LENID: INFO区ASCII字节数
	// LCHKSUM: D11D10D9D8 + D7D6D5D4 + D3D2D1D0，模16取反加1
	d11_8 := (lenid >> 8) & 0x0F
	d7_4 := (lenid >> 4) & 0x0F
	d3_0 := lenid & 0x0F
	sum := d11_8 + d7_4 + d3_0
	lchksum := (^sum + 1) & 0x0F
	length := (uint16(lchksum) << 12) | (lenid & 0x0FFF)
	// 转为2字节，再转ASCII
	return bytesToASCIIHex([]byte{byte(length >> 8), byte(length & 0xFF)})
}

// ASCII字符转字节
func asciiToByte(hi, lo byte) (byte, error) {
	var h, l byte
	switch {
	case hi >= '0' && hi <= '9':
		h = hi - '0'
	case hi >= 'A' && hi <= 'F':
		h = hi - 'A' + 10
	case hi >= 'a' && hi <= 'f':
		h = hi - 'a' + 10
	default:
		return 0, errors.New("invalid hex character")
	}
	switch {
	case lo >= '0' && lo <= '9':
		l = lo - '0'
	case lo >= 'A' && lo <= 'F':
		l = lo - 'A' + 10
	case lo >= 'a' && lo <= 'f':
		l = lo - 'a' + 10
	default:
		return 0, errors.New("invalid hex character")
	}
	return (h << 4) | l, nil
}

// ASCII字符串转字节数组
func asciiToBytes(ascii string) ([]byte, error) {
	if len(ascii)%2 != 0 {
		return nil, errors.New("ascii string length must be even")
	}
	result := make([]byte, len(ascii)/2)
	for i := 0; i < len(ascii); i += 2 {
		b, err := asciiToByte(ascii[i], ascii[i+1])
		if err != nil {
			return nil, err
		}
		result[i/2] = b
	}
	return result, nil
}

// 验证LENGTH字段的LCHKSUM
func verifyLCHKSUM(length [2]byte) error {
	// 提取LENID和LCHKSUM
	lenid := uint16(length[0]&0x0F)<<8 | uint16(length[1])
	lchksum := (length[0] >> 4) & 0x0F

	// 计算LCHKSUM
	d11_8 := (lenid >> 8) & 0x0F
	d7_4 := (lenid >> 4) & 0x0F
	d3_0 := lenid & 0x0F
	sum := d11_8 + d7_4 + d3_0
	expectedLCHKSUM := (^sum + 1) & 0x0F

	if lchksum != byte(expectedLCHKSUM) {
		return errors.New("LCHKSUM verification failed")
	}
	return nil
}

// Decode 解码协议帧
func Decode(adu []byte) (pdu *ProtocolDataUnit, err error) {
	if len(adu) < 9 {
		return nil, errors.New("frame too short")
	}

	// 检查SOI和EOI
	if adu[0] != SOI || adu[len(adu)-1] != EOI {
		return nil, errors.New("invalid SOI or EOI")
	}

	// 提取ASCII部分（去掉SOI、EOI、CHKSUM）
	asciiPart := string(adu[1 : len(adu)-5])

	// 验证CHKSUM
	receivedCHKSUM := string(adu[len(adu)-5 : len(adu)-1])
	calculatedCHKSUM := CalcCHKSUM(asciiPart)
	if receivedCHKSUM != calculatedCHKSUM {
		return nil, errors.New("CHKSUM verification failed")
	}

	// 解析ASCII部分
	// 格式：VER(2) + ADR(2) + CID1(2) + CID2(2) + LENGTH(4) + INFO(2*lenid)
	if len(asciiPart) < 12 {
		return nil, errors.New("invalid frame format")
	}

	// 解析基本字段
	ver, err := asciiToByte(asciiPart[0], asciiPart[1])
	if err != nil {
		return nil, err
	}

	addr, err := asciiToByte(asciiPart[2], asciiPart[3])
	if err != nil {
		return nil, err
	}

	cid1, err := asciiToByte(asciiPart[4], asciiPart[5])
	if err != nil {
		return nil, err
	}
	// cid2 or RTN
	cid2, err := asciiToByte(asciiPart[6], asciiPart[7])
	if err != nil {
		return nil, err
	}

	//Check RTN Code
	err = mapRTNCodeToError(cid2)
	if err != nil {
		return nil, err
	}

	// 解析LENGTH字段
	lengthBytes, err := asciiToBytes(asciiPart[8:12])
	if err != nil {
		return nil, err
	}
	var length [2]byte
	copy(length[:], lengthBytes)

	// 验证LCHKSUM
	if err := verifyLCHKSUM(length); err != nil {
		return nil, err
	}

	// 提取LENID
	lenid := uint16(length[0]&0x0F)<<8 | uint16(length[1])

	// 解析INFO字段
	var info []byte
	if lenid > 0 {
		if len(asciiPart) < int(12+lenid) {
			return nil, errors.New("frame length mismatch")
		}
		infoASCII := asciiPart[12 : 12+lenid]
		info, err = asciiToBytes(infoASCII)
		if err != nil {
			return nil, err
		}
	}

	// 提取CHKSUM
	chksumBytes, err := asciiToBytes(receivedCHKSUM)
	if err != nil {
		return nil, err
	}
	var chksum [2]byte
	copy(chksum[:], chksumBytes)

	return &ProtocolDataUnit{
		Ver:  ver,
		Addr: addr,
		Cid1: cid1,
		Cid2: cid2,
		Info: info,
	}, nil
}

// Encode 编码协议帧
func Encode(pdu *ProtocolDataUnit) (adu []byte, err error) {
	// 1. 组装主字段
	frame := []byte{pdu.Ver, pdu.Addr, pdu.Cid1, pdu.Cid2}

	// 2. 组装INFO区
	infoASCII := bytesToASCIIHex(pdu.Info)
	lenid := uint16(len(infoASCII))

	// 3. 组装LENGTH字段
	lengthASCII := calcLENGTH(lenid)

	// 4. 拼接全部ASCII区
	mainASCII := append(bytesToASCIIHex(frame), lengthASCII...)
	mainASCII = append(mainASCII, infoASCII...)

	// 5. 计算CHKSUM
	chksum := CalcCHKSUM(string(mainASCII))

	// 6. 拼接最终帧
	final := []byte{SOI}
	final = append(final, mainASCII...)
	final = append(final, []byte(chksum)...)
	final = append(final, EOI)

	return final, nil
}
