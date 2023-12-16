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
	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:])
	// 响应长度
	length := headerLength + bodyLength
	data := make([]byte, length)
	_, err = conn.Read(data[numOfLengthBytes:])
	copy(data[:numOfLengthBytes], lenBs)
	return data, err
}
