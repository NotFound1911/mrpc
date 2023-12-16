package mrpc

import (
	"context"
)

type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) // 一个GetById方法
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
func (u *UserServiceServer) Name() string {
	return "user-service"
}
