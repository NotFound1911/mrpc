package message

import (
	"bytes"
	"encoding/binary"
)

const (
	nameSeparator = '\n'
	metaSeparator = '\r'
)

type Request struct {
	// 头部
	HeadLength uint32 // 消息长度
	BodyLength uint32 // 协议版本
	RequestID  uint32 // 消息ID
	Version    uint8  // 版本
	Compresser uint8  // 压缩算法
	Serializer uint8  // 序列化协议
	// 服务名和方法名
	ServiceName string
	MethodName  string
	// 扩展字段，由于传递自定义元数据
	Meta map[string]string
	// 协议体
	Data []byte
	Arg  []byte
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLength+req.BodyLength)
	// 1.写入头部长度
	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	// 2.写入body长度
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	// 3.写入request id
	binary.BigEndian.PutUint32(bs[8:12], req.RequestID)
	// 4.Version Compresser Serializer
	bs[12] = req.Version
	bs[13] = req.Compresser
	bs[14] = req.Serializer
	// 5.写入ServiceName
	cur := bs[15:]
	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	cur[0] = nameSeparator
	cur = cur[1:]
	// 6.写入MethodName
	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]
	cur[0] = nameSeparator
	cur = cur[1:]
	// meta
	for k, v := range req.Meta {
		copy(cur, k)
		cur = cur[len(k):]
		cur[0] = metaSeparator
		cur = cur[1:]
		copy(cur, v)
		cur = cur[len(v):]
		cur[0] = nameSeparator
		cur = cur[1:]
	}
	// data
	copy(cur, req.Data)
	return bs
}
func DecodeReq(data []byte) *Request {
	req := &Request{}
	// 1.头部长度
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2.body长度
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3.request id
	req.RequestID = binary.BigEndian.Uint32(data[8:12])
	// 4.Version Compresser Serializer
	req.Version = data[12]
	req.Compresser = data[13]
	req.Serializer = data[14]
	// 5.ServiceName
	header := data[15:req.HeadLength]
	index := bytes.IndexByte(header, nameSeparator)
	req.ServiceName = string(header[:index])
	header = header[index+1:]
	// 6.MethodName
	index = bytes.IndexByte(header, nameSeparator)
	req.MethodName = string(header[:index])
	header = header[index+1:]
	// meta
	index = bytes.IndexByte(header, nameSeparator)
	if index != -1 {
		meta := make(map[string]string, 8)
		for index != -1 {
			pair := header[:index]
			pairIndex := bytes.IndexByte(pair, metaSeparator)
			key := string(pair[:pairIndex])
			val := string(pair[pairIndex+1:])
			meta[key] = val

			header = header[index+1:]
			index = bytes.IndexByte(header, nameSeparator)
		}
		req.Meta = meta
	}
	if req.BodyLength != 0 {
		req.Data = data[req.HeadLength:]
	}

	return req
}
