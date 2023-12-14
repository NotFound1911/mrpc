package message

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
