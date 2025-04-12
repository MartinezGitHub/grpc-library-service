package controller

import (
	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetAuthorBooks(req *library.GetAuthorBooksRequest, stream library.Library_GetAuthorBooksServer) error {
	if err := req.ValidateAll(); err != nil {
		i.logger.Error("GetAuthorBooks validation failed: ", zap.Error(err))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	authorChan, errChan := i.authorUseCase.StreamBooksForAuthor(stream.Context(), i.logger, req.GetAuthorId())
	for {
		select {
		case err := <-errChan:
			return i.convertErr(err)
		case book, ok := <-authorChan:
			if !ok {
				return nil
			}
			i.logger.Info("took book " + book.ID + " to chan")
			err := stream.Send(&library.Book{
				Id:       book.ID,
				Name:     book.Name,
				AuthorId: book.AuthorIDs},
			)
			if err != nil {
				return i.convertErr(err)
			}
			i.logger.Info("successfully send book " + book.ID + " to stream")
		case <-stream.Context().Done():
			return i.convertErr(stream.Context().Err())
		}
	}
}
