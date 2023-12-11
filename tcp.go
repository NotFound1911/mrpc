package mrpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	// 响应长度
	length := binary.BigEndian.Uint64(lenBs)
	data := make([]byte, length)
	_, err = conn.Read(data)
	return data, err
}

func EncodeMsg(data []byte) []byte {
	respLen := len(data)

	// 构建响应数据
	res := make([]byte, respLen+numOfLengthBytes)
	// step1.写入长度
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
	// step2.写入数据
	copy(res[numOfLengthBytes:], data)
	return res
}
