package mrpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Server struct {
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service, 16),
	}
}
func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if errConn := s.handleConn(conn); errConn != nil {
				conn.Close()
			}
		}()
	}
}

// 请求组成:
// part1. 长度字段，用固定字节表示
// part2. 请求数据
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		// lenBs 长度字段的字节表示
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}
		// 获取消息长度
		length := binary.BigEndian.Uint64(lenBs)

		reqBs := make([]byte, length)
		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}

		respData, err := s.handleMsg(reqBs)
		if err != nil {
			return err
		}
		respLen := len(respData)

		// 构建响应数据
		res := make([]byte, respLen+numOfLengthBytes)
		// step1.写入长度
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
		// step2.写入数据
		copy(res[numOfLengthBytes:], respData)

		if _, err = conn.Write(res); err != nil {
			return err
		}
	}
	return nil
}
func (s *Server) handleMsg(reqData []byte) ([]byte, error) {
	// 还原调用信息
	req := &Request{}
	err := json.Unmarshal(reqData, req)
	if err != nil {
		return nil, err
	}
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("调用的服务不存在")
	}
	val := reflect.ValueOf(service)
	method := val.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)

	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	err = json.Unmarshal(req.Arg, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)
	// results[0] 返回值
	// results[1] error
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}
	resp, err := json.Marshal(results[0].Interface())
	return resp, err
}
