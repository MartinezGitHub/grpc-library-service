package controller

import (
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	library2 "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/mocks"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var InvalidAuthorsIDs = strings.Split(strings.Repeat("1,", 101)[:len(strings.Repeat("1,", 101))-1], ",")
var CreationTime = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

func TestAddBook(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		req          *library2.AddBookRequest
		resp         *library2.AddBookResponse
		expectedBook entity.Book
		expectedErr  error
		errStatus    codes.Code
	}{
		{
			name: "TestAddBook",
			req: &library2.AddBookRequest{
				Name:      "Test Book",
				AuthorIds: []string{"1", "2"},
			},
			resp: &library2.AddBookResponse{
				Book: &library2.Book{
					Id:        "1",
					Name:      "Book 1",
					AuthorId:  []string{"1", "2"},
					CreatedAt: timestamppb.New(CreationTime),
					UpdatedAt: timestamppb.New(CreationTime),
				},
			},
			expectedBook: entity.Book{
				ID:        "1",
				Name:      "Book 1",
				AuthorIDs: []string{"1", "2"},
				CreatedAt: CreationTime,
				UpdatedAt: CreationTime,
			},
			expectedErr: nil,
			errStatus:   codes.OK,
		},
		{
			name: "ErrAuthorNotFound",
			req: &library2.AddBookRequest{
				Name:      "Test Book",
				AuthorIds: []string{"1", "2"},
			},
			resp:         nil,
			expectedBook: entity.Book{},
			expectedErr:  entity.ErrAuthorNotFound,
			errStatus:    codes.NotFound,
		},
		{
			name: "ErrBookAlreadyExists",
			req: &library2.AddBookRequest{
				Name:      "Test Book",
				AuthorIds: []string{"1", "2"},
			},
			resp:         nil,
			expectedBook: entity.Book{},
			expectedErr:  entity.ErrBookAlreadyExists,
			errStatus:    codes.AlreadyExists,
		},
		{
			name: "AuthorIDsValidateFail",
			req: &library2.AddBookRequest{
				Name:      "Test Book",
				AuthorIds: InvalidAuthorsIDs,
			},
			resp:         nil,
			expectedBook: entity.Book{},
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can not initialize logger: %s", err)
			}
			ctx := context.Background()

			mockBooksUseCase := mocks.NewMockBooksUseCase(ctrl)
			controller := New(logger, mockBooksUseCase, nil)

			if tc.errStatus != codes.InvalidArgument {
				mockBooksUseCase.EXPECT().RegisterBook(ctx, logger, tc.req.GetName(), tc.req.GetAuthorIds()).Return(tc.expectedBook, tc.expectedErr)
			}

			response, err := controller.AddBook(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name        string
		req         *library2.ChangeAuthorInfoRequest
		resp        *library2.ChangeAuthorInfoResponse
		expectedErr error
		errStatus   codes.Code
	}{
		{
			name: "TestChangeAuthorInfo",
			req: &library2.ChangeAuthorInfoRequest{
				Id:   "550e8400-e29b-41d4-a716-446655440000",
				Name: "newName",
			},
			resp:        &library2.ChangeAuthorInfoResponse{},
			expectedErr: nil,
			errStatus:   codes.OK,
		},
		{
			name: "ErrAuthorNotFound",
			req: &library2.ChangeAuthorInfoRequest{
				Id:   "550e8400-e29b-41d4-a716-446655440000",
				Name: "newName",
			},
			resp:        nil,
			expectedErr: entity.ErrAuthorNotFound,
			errStatus:   codes.NotFound,
		},
		{
			name: "IDValidateFail",
			req: &library2.ChangeAuthorInfoRequest{
				Id:   "1",
				Name: "newName",
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
		},
		{
			name: "NameValidateFail",
			req: &library2.ChangeAuthorInfoRequest{
				Id:   "1",
				Name: "@navi",
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)

			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can not initialize logger: %s", err)
			}

			controller := New(logger, nil, mockAuthorUseCase)
			if tc.errStatus != codes.InvalidArgument {
				mockAuthorUseCase.EXPECT().ChangeAuthorInfo(ctx, logger, tc.req.GetId(), tc.req.GetName()).Return(tc.expectedErr)
			}

			response, err := controller.ChangeAuthorInfo(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name           string
		req            *library2.GetAuthorInfoRequest
		resp           *library2.GetAuthorInfoResponse
		expectedAuthor entity.Author
		expectedErr    error
		errStatus      codes.Code
	}{
		{
			name: "TestGetAuthorInfo",
			req: &library2.GetAuthorInfoRequest{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			resp: &library2.GetAuthorInfoResponse{
				Id:   "550e8400-e29b-41d4-a716-446655440000",
				Name: "Author 1",
			},
			expectedAuthor: entity.Author{
				ID:   "550e8400-e29b-41d4-a716-446655440000",
				Name: "Author 1",
			},
			expectedErr: nil,
			errStatus:   codes.OK,
		},
		{
			name: "ErrAuthorNotFound",
			req: &library2.GetAuthorInfoRequest{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			resp:           nil,
			expectedAuthor: entity.Author{},
			expectedErr:    entity.ErrAuthorNotFound,
			errStatus:      codes.NotFound,
		},
		{

			name: "IDValidateFail",
			req: &library2.GetAuthorInfoRequest{
				Id: "1",
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)

			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can not initialize logger: %s", err)
			}

			controller := New(logger, nil, mockAuthorUseCase)
			if tc.errStatus != codes.InvalidArgument {
				mockAuthorUseCase.EXPECT().GetAuthorInfo(ctx, logger, tc.req.GetId()).Return(tc.expectedAuthor.Name, tc.expectedErr)
			}

			response, err := controller.GetAuthorInfo(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestGetBookInfo(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		req          *library2.GetBookInfoRequest
		resp         *library2.GetBookInfoResponse
		expectedBook entity.Book
		expectedErr  error
		errStatus    codes.Code
	}{
		{
			name: "TestGetBookInfo",
			req: &library2.GetBookInfoRequest{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			resp: &library2.GetBookInfoResponse{
				Book: &library2.Book{
					Id:        "550e8400-e29b-41d4-a716-446655440000",
					Name:      "Book 1",
					AuthorId:  []string{"123e4567-e89b-12d3-a456-426614174000"},
					CreatedAt: timestamppb.New(CreationTime),
					UpdatedAt: timestamppb.New(CreationTime),
				},
			},
			expectedBook: entity.Book{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Book 1",
				AuthorIDs: []string{"123e4567-e89b-12d3-a456-426614174000"},
				CreatedAt: CreationTime,
				UpdatedAt: CreationTime,
			},
			expectedErr: nil,
			errStatus:   codes.OK,
		},
		{
			name: "ErrBookNotFound",
			req: &library2.GetBookInfoRequest{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			resp:         nil,
			expectedBook: entity.Book{},
			expectedErr:  entity.ErrBookNotFound,
			errStatus:    codes.NotFound,
		},
		{
			name: "IDValidateFail",
			req: &library2.GetBookInfoRequest{
				Id: "1",
			},
			resp:         nil,
			expectedBook: entity.Book{},
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can not initialize logger: %s", err)
			}

			mockBookUseCase := mocks.NewMockBooksUseCase(ctrl)
			controller := New(logger, mockBookUseCase, nil)
			if tc.errStatus != codes.InvalidArgument {
				mockBookUseCase.EXPECT().GetBook(ctx, logger, tc.req.GetId()).Return(tc.expectedBook, tc.expectedErr)
			}

			response, err := controller.GetBookInfo(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name             string
		req              *library2.RegisterAuthorRequest
		resp             *library2.RegisterAuthorResponse
		expectedAuthorID string
		expectedErr      error
		errStatus        codes.Code
	}{
		{
			name: "TestRegisterAuthor",
			req: &library2.RegisterAuthorRequest{
				Name: "Author 1",
			},
			resp: &library2.RegisterAuthorResponse{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedAuthorID: "550e8400-e29b-41d4-a716-446655440000",
			expectedErr:      nil,
			errStatus:        codes.OK,
		},
		{
			name: "ErrAuthorAlreadyExists",
			req: &library2.RegisterAuthorRequest{
				Name: "Author 1",
			},
			resp:             nil,
			expectedAuthorID: "",
			expectedErr:      entity.ErrAuthorAlreadyExists,
			errStatus:        codes.AlreadyExists,
		},
		{
			name: "NameValidateFail",
			req: &library2.RegisterAuthorRequest{
				Name: "$#@",
			},
			resp:             nil,
			expectedAuthorID: "",
			expectedErr:      nil,
			errStatus:        codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)

			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can not initialize logger: %s", err)
			}

			controller := New(logger, nil, mockAuthorUseCase)
			if tc.errStatus != codes.InvalidArgument {
				mockAuthorUseCase.EXPECT().RegisterAuthor(ctx, logger, tc.req.GetName()).Return(tc.expectedAuthorID, tc.expectedErr)
			}

			response, err := controller.RegisterAuthor(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name        string
		req         *library2.UpdateBookRequest
		resp        *library2.UpdateBookResponse
		expectedErr error
		errStatus   codes.Code
	}{
		{
			name: "TestUpdateBook",
			req: &library2.UpdateBookRequest{
				Id:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Author 1",
				AuthorIds: []string{"123e4567-e89b-12d3-a456-426614174000"},
			},
			resp:        &library2.UpdateBookResponse{},
			expectedErr: nil,
			errStatus:   codes.OK,
		},
		{
			name: "ErrAuthorNotFound",
			req: &library2.UpdateBookRequest{
				Id:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Author 1",
				AuthorIds: []string{"123e4567-e89b-12d3-a456-426614174000"},
			},
			resp:        nil,
			expectedErr: entity.ErrAuthorNotFound,
			errStatus:   codes.NotFound,
		},
		{
			name: "IDValidateFail",
			req: &library2.UpdateBookRequest{
				Id:        "1",
				Name:      "Author 1",
				AuthorIds: []string{"123e4567-e89b-12d3-a456-426614174000"},
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
		},
		{
			name: "AuthorIDsValidateFail",
			req: &library2.UpdateBookRequest{
				Id:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Author 1",
				AuthorIds: InvalidAuthorsIDs,
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockBookUseCase := mocks.NewMockBooksUseCase(ctrl)

			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can not initialize logger: %s", err)
			}

			controller := New(logger, mockBookUseCase, nil)
			if tc.errStatus != codes.InvalidArgument {
				mockBookUseCase.EXPECT().UpdateBook(ctx, logger, tc.req.GetId(),
					tc.req.GetName(),
					tc.req.GetAuthorIds()).Return(tc.expectedErr)
			}

			response, err := controller.UpdateBook(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name        string
		req         *library2.GetAuthorBooksRequest
		books       []entity.Book
		expectedErr error
		errStatus   codes.Code
	}{
		{
			name: "TestGetAuthorBooks",
			req: &library2.GetAuthorBooksRequest{
				AuthorId: "550e8400-e29b-41d4-a716-446655440000",
			},
			books: []entity.Book{
				{
					ID:        "book1",
					Name:      "Book 1",
					AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
				},
				{
					ID:        "book2",
					Name:      "Book 2",
					AuthorIDs: []string{"550e8400-e29b-41d4-a716-446655440000"},
				},
			},
			expectedErr: nil,
			errStatus:   codes.OK,
		},
		{
			name: "AuthorIDsValidateFail",
			req: &library2.GetAuthorBooksRequest{
				AuthorId: "1",
			},
			books:       nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			mockStream := mocks.NewMockLibrary_GetAuthorBooksServer(ctrl)
			logger := zap.NewNop()
			controller := New(logger, nil, mockAuthorUseCase)

			mockStream.EXPECT().Context().Return(ctx).AnyTimes()

			if tc.errStatus == codes.InvalidArgument {
				err := controller.GetAuthorBooks(tc.req, mockStream)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.errStatus, s.Code())
				return
			}

			bookCh := make(chan entity.Book, len(tc.books))
			errCh := make(chan error, 1)

			mockAuthorUseCase.EXPECT().
				StreamBooksForAuthor(ctx, logger, tc.req.GetAuthorId()).
				Return((<-chan entity.Book)(bookCh), (<-chan error)(errCh))

			for _, book := range tc.books {
				bookCh <- book
			}
			close(bookCh)

			for _, book := range tc.books {
				mockStream.EXPECT().
					Send(&library2.Book{
						Id:       book.ID,
						Name:     book.Name,
						AuthorId: book.AuthorIDs,
					}).
					Return(nil)
			}

			err := controller.GetAuthorBooks(tc.req, mockStream)

			if tc.expectedErr != nil {
				require.Error(t, err)
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.errStatus, s.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
