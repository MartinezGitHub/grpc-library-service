package controller

import (
	"context"

	"go.uber.org/zap"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetBookInfo(ctx context.Context, req *library.GetBookInfoRequest) (*library.GetBookInfoResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Error("GetBookInfo validation failed: ", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	i.logger.Debug("Try to get info for book: " + req.GetId())
	resp, err := i.booksUseCase.GetBook(ctx, i.logger, req.GetId())
	if err != nil {
		i.logger.Error("GetBookInfo failed: ", zap.Error(err))
		return nil, i.convertErr(err)
	}

	i.logger.Info("Successfully get info for book: " + req.GetId())

	return resp, nil
}
