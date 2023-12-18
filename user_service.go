package mrpc

import (
	"context"
	"github.com/NotFound1911/mrpc/internal/proto/gen"
	"testing"
	"time"
)

type UserService struct {
	GetById      func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) // 一个GetById方法
	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}
type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}
func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: u.Msg,
		},
	}, u.Err
}
func (u *UserServiceServer) Name() string {
	return "user-service"
}

type UserServiceServerTimeout struct {
	t     *testing.T
	sleep time.Duration
	Err   error
	Msg   string
}

func (u *UserServiceServerTimeout) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	if _, ok := ctx.Deadline(); !ok {
		u.t.Fatal("未设置超时时间")
	}
	time.Sleep(u.sleep)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}
func (u *UserServiceServerTimeout) Name() string {
	return "user-service"
}
