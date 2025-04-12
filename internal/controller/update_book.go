package controller

import (
	"context"

	"go.uber.org/zap"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) UpdateBook(ctx context.Context, req *library.UpdateBookRequest) (*library.UpdateBookResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Error("UpdateBook validation failed: ", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	i.logger.Debug("Try to update book: " + req.GetId())
	err := i.booksUseCase.UpdateBook(ctx, i.logger, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		i.logger.Error("UpdateBook failed: ", zap.Error(err))
		return nil, i.convertErr(err)
	}

	i.logger.Info("Success to update book: " + req.GetId())
	return &library.UpdateBookResponse{}, nil
}
