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
	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}
		// 本地调用捕捉到的地方
		fn := func(args []reflect.Value) (results []reflect.Value) {
			ctx := args[0].Interface().(context.Context)
			retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
			reqData, err := json.Marshal(args[1].Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			req := &Request{
				ServiceName: service.Name(),
				MethodName:  fieldTyp.Name,
				Arg:         reqData,
			}

			// 发起调用
			resp, err := p.Invoke(ctx, req)
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			err = json.Unmarshal(resp.Data, retVal.Interface())
			if err != nil {
				return []reflect.Value{retVal, reflect.ValueOf(err)}
			}
			return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
		}
		// 设置值
		fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
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
