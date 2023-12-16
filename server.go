package mrpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NotFound1911/mrpc/message"
	"net"
	"reflect"
)

type Server struct {
	services map[string]reflectionStub
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]reflectionStub, 16),
	}
}
func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
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
func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("调用的服务不存在")
	}
	resp := &message.Response{
		RequestID:  req.RequestID,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}
	respData, err := service.invoke(ctx, req.MethodName, req.Data)
	resp.Data = respData
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// 请求组成:
// part1. 长度字段，用固定字节表示
// part2. 请求数据
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}
		// 还原调用信息
		req := message.DecodeReq(reqBs)
		if err != nil {
			return err
		}
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// 处理业务 error
			resp.Error = []byte(err.Error())
		}
		resp.CalHeaderLength()
		resp.CalBodyLength()

		if _, err = conn.Write(message.EncodeResp(resp)); err != nil {
			return err
		}
	}
	return nil
}

type reflectionStub struct {
	s     Service
	value reflect.Value
}

func (s *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// 反射找到方法 并执行调用
	// s.value是通过反射保存的结构体 MethodByName是结构体的方法
	method := s.value.MethodByName(methodName)
	in := make([]reflect.Value, 2)

	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	// 解析请求
	err := json.Unmarshal(data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	// 第二个参数是根据方法的输入参数类型动态创建的指针类型的值，它会被用来接收传入的数据
	in[1] = inReq
	results := method.Call(in) // 调用结构体方法
	// results[0] 返回值
	// results[1] error
	if results[1].Interface() != nil {
		err = results[1].Interface().(error)
	}
	var res []byte
	if results[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = json.Marshal(results[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, err
}
