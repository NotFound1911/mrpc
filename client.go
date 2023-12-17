package mrpc

import (
	"context"
	"errors"
	"github.com/NotFound1911/mrpc/message"
	"github.com/NotFound1911/mrpc/serialize"
	"github.com/NotFound1911/mrpc/serialize/json"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"time"
)

const numOfLengthBytes = 8

// InitService 为GetById之类的函数类型字段赋值
func (c *Client) InitService(service Service) error {
	return setFuncField(service, c, c.serializer)
}
func setFuncField(service Service, p Proxy, s serialize.Serializer) error {
	if service == nil {
		return errors.New("mrpc: 不支持nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("mrpc: 只支持指向结构体的一级指针")
	}
	// 获取指针指向的实际数值和类型
	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	// numField 为Proxy方法数量
	for i := 0; i < numField; i++ {
		// 获取字段的类型和值
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}
		// 本地调用捕捉到的地方
		fn := func(args []reflect.Value) (results []reflect.Value) {
			ctx := args[0].Interface().(context.Context)
			// retVal 是一个指向输出参数类型的新指针，用于存储远程调用的结果
			retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
			// 将请求数据序列化为
			reqData, err := s.Encode(args[1].Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			// 创建Request对象
			// 根据函数字段构建请求
			req := &message.Request{
				ServiceName: service.Name(),
				MethodName:  fieldTyp.Name,
				Data:        reqData,
				Serializer:  s.Code(),
			}
			req.CalHeaderLen()
			req.CalBodyLen()
			// 发起调用，调用代理对象的Invoke方法
			resp, err := p.Invoke(ctx, req)
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			var retErr error
			if len(resp.Error) > 0 {
				retErr = errors.New(string(resp.Error))
			}
			if len(resp.Data) > 0 {
				// 将响应数据解析为目标结构体并赋值给retVal
				err = s.Decode(resp.Data, retVal.Interface())
				if err != nil {
					// 反序列化的err
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
			}
			var retErrVal reflect.Value
			if retErr == nil {
				retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
			} else {
				retErrVal = reflect.ValueOf(retErr)
			}
			return []reflect.Value{retVal, retErrVal}
		}
		// 使用反射创建一个函数值
		fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
		// 设置字段的值为创建的函数值
		fieldVal.Set(fnVal)
	}
	return nil
}

type Client struct {
	pool       pool.Pool
	serializer serialize.Serializer
}
type ClientOption func(client *Client)

func ClientWithSerializer(sl serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = sl
	}
}
func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap:  5,
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Minute,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	res := &Client{
		pool:       p,
		serializer: &json.Serializer{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}
func (c Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	// 把请求发送至服务端
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		c.pool.Put(val)
	}()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	return ReadMsg(conn)
}
