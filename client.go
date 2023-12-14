package mrpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"time"
)

const numOfLengthBytes = 8

// InitClientProxy 为GetById之类的函数类型字段赋值
func InitClientProxy(addr string, service Service) error {
	client, err := NewClient(addr)
	if err != nil {
		return err
	}
	return setFuncField(service, client)
}
func setFuncField(service Service, p Proxy) error {
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
			// 将请求数据序列化为JSON
			reqData, err := json.Marshal(args[1].Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			// 创建Request对象
			// 根据函数字段构建请求
			req := &Request{
				ServiceName: service.Name(),
				MethodName:  fieldTyp.Name,
				Arg:         reqData,
			}

			// 发起调用，调用代理对象的Invoke方法
			resp, err := p.Invoke(ctx, req)
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			// 将响应数据解析为目标结构体并赋值给retVal
			err = json.Unmarshal(resp.Data, retVal.Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
		}
		// 使用反射创建一个函数值
		fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
		// 设置字段的值为创建的函数值
		fieldVal.Set(fnVal)
	}
	return nil
}

type Client struct {
	pool pool.Pool
}

func NewClient(addr string) (*Client, error) {
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
	return &Client{
		pool: p,
	}, nil
}
func (c Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	// 把请求发送至服务端
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: resp,
	}, nil
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
	req := EncodeMsg(data)
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}
	return ReadMsg(conn)
}
