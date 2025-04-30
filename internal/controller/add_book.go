package controller

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project/library/generated/api/library"
)

func (i *implementation) AddBook(ctx context.Context, req *library.AddBookRequest) (*library.AddBookResponse, error) {
	if err := req.ValidateAll(); err != nil {
		i.logger.Error("AddBook: validation failed: ", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	i.logger.Debug("Try to register book: " + req.GetName())
	resp, err := i.booksUseCase.RegisterBook(ctx, i.logger, req.GetName(), req.GetAuthorIds())

	if err != nil {
		i.logger.Error("UseCase error:", zap.Error(err))
		return nil, i.convertErr(err)
	}
	i.logger.Info("Successfully register book: " + req.GetName())
	return resp, err
}
