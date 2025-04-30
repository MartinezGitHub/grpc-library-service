package controller

import (
	"context"

	"go.uber.org/zap"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) RegisterAuthor(ctx context.Context, req *library.RegisterAuthorRequest) (*library.RegisterAuthorResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Error("RegisterAuthor validation failed: ", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	i.logger.Debug("Try to register author with name: " + req.GetName())
	resp, err := i.authorUseCase.RegisterAuthor(ctx, i.logger, req.GetName())

	if err != nil {
		i.logger.Error("Failed to register author with name: " + req.GetName())
		return nil, i.convertErr(err)
	}
	i.logger.Info("Successfully register author with name: " + req.GetName())
	return resp, nil
}
