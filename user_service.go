package mrpc

import (
	"context"
	"log"
)

type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
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
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println("req:", req)
	return &GetByIdResp{
		Msg: "test",
	}, nil
}
func (u *UserServiceServer) Name() string {
	return "user-service"
}
