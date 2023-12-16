package mrpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log("err:", err)
	}()
	time.Sleep(time.Second * 3)
	usClient := &UserService{} // 客户端服务
	err := InitClientProxy(":8081", usClient)
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
