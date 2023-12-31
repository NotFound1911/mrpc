package mrpc

import (
	"context"
	"errors"
	"github.com/NotFound1911/mrpc/internal/proto/gen"
	"github.com/NotFound1911/mrpc/serialize/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
// cd internal/proto
// protoc --go_out=. user.proto
func TestInitServiceProto(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerializer(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log("err:", err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{} // 客户端服务
	//client, err := NewClient(":8081") // json 协议
	client, err := NewClient(":8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("test error")
			},
			wantResp: &GetByIdResp{},
			wantErr:  errors.New("test error"),
		},
		{
			name: "both",
			mock: func() {
				service.Msg = "hello world"
				service.Err = errors.New("test error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello world",
			},
			wantErr: errors.New("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			//resp, err := usClient.GetById(context.Background(), &GetByIdReq{
			//	Id: 123,
			//})
			resp, err := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{
				Id: 123,
			})
			assert.Equal(t, tc.wantErr, err)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}
}

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log("err:", err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}        // 客户端服务
	client, err := NewClient(":8081") // json 协议
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("test error")
			},
			wantResp: &GetByIdResp{},
			wantErr:  errors.New("test error"),
		},
		{
			name: "both",
			mock: func() {
				service.Msg = "hello world"
				service.Err = errors.New("test error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello world",
			},
			wantErr: errors.New("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetById(context.Background(), &GetByIdReq{
				Id: 123,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestOneway(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log("err:", err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}        // 客户端服务
	client, err := NewClient(":8081") // json 协议
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "oneway",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "test"
			},
			wantResp: &GetByIdResp{},
			wantErr:  errors.New("mrpc: oneway调用，不应该处理任何结果"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx := CtxWithOneway(context.Background())
			resp, err := usClient.GetById(ctx, &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestTimeout(t *testing.T) {
	server := NewServer()
	service := &UserServiceServerTimeout{t: t}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log("err:", err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{}        // 客户端服务
	client, err := NewClient(":8081") // json 协议
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct {
		name string
		mock func() context.Context

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "timeout",
			mock: func() context.Context {
				service.Msg = "test"
				service.Err = errors.New("mock error")
				// 服务睡眠2s
				// 超时设置了1s，客户端预期得到超时响应
				service.sleep = 2 * time.Second
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			},
			wantResp: &GetByIdResp{},
			wantErr:  context.DeadlineExceeded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, er := usClient.GetById(tc.mock(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}
