package controller

import (
	"context"

	"go.uber.org/zap"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetAuthorInfo(ctx context.Context, req *library.GetAuthorInfoRequest) (*library.GetAuthorInfoResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Error("GetAuthorInfo validation failed: ", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	i.logger.Debug("Try to get info for author: " + req.GetId())
	resp, err := i.authorUseCase.GetAuthorInfo(ctx, i.logger, req.GetId())
	if err != nil {
		i.logger.Error("GetAuthorInfo failed: ", zap.Error(err))
		return nil, i.convertErr(err)
	}
	i.logger.Info("Successfully get info for author: " + req.GetId())
	return resp, nil
	//return &library.GetAuthorInfoResponse{
	//	Id:   req.GetId(),
	//	Name: name,
	//}, nil
}
