package mrpc

import (
	"context"
	"github.com/NotFound1911/mrpc/internal/proto/gen"
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
