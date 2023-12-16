package message

import "encoding/binary"

type Response struct {
	// 头部
	HeadLength uint32 // 消息长度
	BodyLength uint32 // 协议版本
	RequestID  uint32 // 消息ID
	Version    uint8  // 版本
	Compresser uint8  // 压缩算法
	Serializer uint8  // 序列化协议
	Error      []byte
	Data       []byte
}

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLength+resp.BodyLength)
	// 1.写入头部长度
	binary.BigEndian.PutUint32(bs[:4], resp.HeadLength)
	// 2.写入body长度
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLength)
	// 3.写入request id
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestID)
	// 4.Version Compresser Serializer
	bs[12] = resp.Version
	bs[13] = resp.Compresser
	bs[14] = resp.Serializer

	cur := bs[15:]
	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]
	copy(cur, resp.Data)

	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}
	// 1.头部长度
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2.body长度
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3.request id
	resp.RequestID = binary.BigEndian.Uint32(data[8:12])
	// 4.Version Compresser Serializer
	resp.Version = data[12]
	resp.Compresser = data[13]
	resp.Serializer = data[14]
	if resp.HeadLength > 15 {
		resp.Error = data[15:resp.HeadLength]
	}
	if resp.BodyLength != 0 {
		resp.Data = data[resp.HeadLength:]
	}
	return resp
}

func (resp *Response) CalHeaderLength() {
	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
