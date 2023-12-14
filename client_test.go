package mrpc

import (
	"context"
	"errors"
	"github.com/NotFound1911/mrpc/message"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

// $GOPATH/bin/mockgen -destination=mock_proxy_test.gen.go -package=mrpc -source=types.go Proxy
func Test_setFuncField(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(controller *gomock.Controller) Proxy
		service Service
		wantErr error
	}{
		{
			name:    "nil",
			service: nil,
			mock: func(controller *gomock.Controller) Proxy {
				return NewMockProxy(controller)
			},
			wantErr: errors.New("mrpc: 不支持nil"),
		},
		{
			name:    "no pointer",
			service: UserService{},
			wantErr: errors.New("mrpc: 只支持指向结构体的一级指针"),
			mock: func(controller *gomock.Controller) Proxy {
				return NewMockProxy(controller)
			},
		},
		{
			name:    "user service",
			service: &UserService{},
			mock: func(controller *gomock.Controller) Proxy {
				p := NewMockProxy(controller)
				p.EXPECT().Invoke(gomock.Any(), &message.Request{
					ServiceName: "user-service",
					MethodName:  "GetById",
					Arg:         []byte(`{"Id":123}`),
				}).Return(&message.Response{}, nil)
				return p
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			err := setFuncField(tc.service, tc.mock(ctrl))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			resp, err := tc.service.(*UserService).GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			t.Log("resp:", resp)
		})
	}
}
