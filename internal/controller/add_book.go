package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	//return &library.AddBookResponse{
	//	Book: &library.Book{
	//		Id:        book.ID,
	//		Name:      book.Name,
	//		AuthorId:  book.AuthorIDs,
	//		CreatedAt: timestamppb.New(book.CreatedAt),
	//		UpdatedAt: timestamppb.New(book.UpdatedAt),
	//	},
	//}, nil
}
