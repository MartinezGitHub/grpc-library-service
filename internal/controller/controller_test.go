package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	library2 "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/dto"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/mocks"
)

var InvalidAuthorsIDs = strings.Split(strings.Repeat("1,", 1005)[:len(strings.Repeat("1,", 1005))-1], ",")
var CreationTime = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

func TestAddBook(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		req          *library2.AddBookRequest
		resp         *library2.AddBookResponse
		expectedResp *library2.AddBookResponse
		expectedErr  error
		errStatus    codes.Code
		shouldCall   bool
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
			expectedResp: &library2.AddBookResponse{
				Book: &library2.Book{
					Id:        "1",
					Name:      "Book 1",
					AuthorId:  []string{"1", "2"},
					CreatedAt: timestamppb.New(CreationTime),
					UpdatedAt: timestamppb.New(CreationTime),
				},
			},
			expectedErr: nil,
			errStatus:   codes.OK,
			shouldCall:  true,
		},
		{
			name: "ErrAuthorNotFound",
			req: &library2.AddBookRequest{
				Name:      "Test Book",
				AuthorIds: []string{"1", "2"},
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  entity.ErrAuthorNotFound,
			errStatus:    codes.NotFound,
			shouldCall:   true,
		},
		{
			name: "AuthorIDsValidateFail",
			req: &library2.AddBookRequest{
				Name:      "Test Book",
				AuthorIds: InvalidAuthorsIDs,
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
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

			if tc.shouldCall {
				fmt.Println("sldf")
				mockBooksUseCase.EXPECT().
					RegisterBook(ctx, logger, tc.req.GetName(), tc.req.GetAuthorIds()).
					Return(tc.expectedResp, tc.expectedErr)
			}

			response, err := controller.AddBook(ctx, tc.req)
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
		expectedResp *library2.GetBookInfoResponse
		expectedErr  error
		errStatus    codes.Code
		shouldCall   bool
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
			expectedResp: &library2.GetBookInfoResponse{
				Book: &library2.Book{
					Id:        "550e8400-e29b-41d4-a716-446655440000",
					Name:      "Book 1",
					AuthorId:  []string{"123e4567-e89b-12d3-a456-426614174000"},
					CreatedAt: timestamppb.New(CreationTime),
					UpdatedAt: timestamppb.New(CreationTime),
				},
			},
			expectedErr: nil,
			errStatus:   codes.OK,
			shouldCall:  true,
		},
		{
			name: "ErrBookNotFound",
			req: &library2.GetBookInfoRequest{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  entity.ErrBookNotFound,
			errStatus:    codes.NotFound,
			shouldCall:   true,
		},
		{
			name: "InvalidBookID",
			req: &library2.GetBookInfoRequest{
				Id: "invalid-id",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
		},
		{
			name: "EmptyBookID",
			req: &library2.GetBookInfoRequest{
				Id: "",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
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

			if tc.shouldCall {
				mockBookUseCase.EXPECT().
					GetBook(ctx, logger, tc.req.GetId()).
					Return(tc.expectedResp, tc.expectedErr)
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
		name         string
		req          *library2.RegisterAuthorRequest
		resp         *library2.RegisterAuthorResponse
		expectedResp *library2.RegisterAuthorResponse
		expectedErr  error
		errStatus    codes.Code
		shouldCall   bool
	}{
		{
			name: "TestRegisterAuthor",
			req: &library2.RegisterAuthorRequest{
				Name: "Author 1",
			},
			resp: &library2.RegisterAuthorResponse{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedResp: &library2.RegisterAuthorResponse{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			expectedErr: nil,
			errStatus:   codes.OK,
			shouldCall:  true,
		},
		{
			name: "ErrAuthorAlreadyExists",
			req: &library2.RegisterAuthorRequest{
				Name: "Author 1",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  entity.ErrAuthorAlreadyExists,
			errStatus:    codes.AlreadyExists,
			shouldCall:   true,
		},
		{
			name: "EmptyName",
			req: &library2.RegisterAuthorRequest{
				Name: "",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
		},
		{
			name: "InvalidNameCharacters",
			req: &library2.RegisterAuthorRequest{
				Name: "Invalid@Name",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
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

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			controller := New(logger, nil, mockAuthorUseCase)

			if tc.shouldCall {
				mockAuthorUseCase.EXPECT().
					RegisterAuthor(ctx, logger, tc.req.GetName()).
					Return(tc.expectedResp, tc.expectedErr)
			}

			response, err := controller.RegisterAuthor(ctx, tc.req)
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
		name         string
		req          *library2.GetAuthorInfoRequest
		resp         *library2.GetAuthorInfoResponse
		expectedResp *library2.GetAuthorInfoResponse
		expectedErr  error
		errStatus    codes.Code
		shouldCall   bool
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
			expectedResp: &library2.GetAuthorInfoResponse{
				Id:   "550e8400-e29b-41d4-a716-446655440000",
				Name: "Author 1",
			},
			expectedErr: nil,
			errStatus:   codes.OK,
			shouldCall:  true,
		},
		{
			name: "ErrAuthorNotFound",
			req: &library2.GetAuthorInfoRequest{
				Id: "550e8400-e29b-41d4-a716-446655440000",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  entity.ErrAuthorNotFound,
			errStatus:    codes.NotFound,
			shouldCall:   true,
		},
		{
			name: "InvalidAuthorID",
			req: &library2.GetAuthorInfoRequest{
				Id: "invalid-id",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
		},
		{
			name: "EmptyAuthorID",
			req: &library2.GetAuthorInfoRequest{
				Id: "",
			},
			resp:         nil,
			expectedResp: nil,
			expectedErr:  nil,
			errStatus:    codes.InvalidArgument,
			shouldCall:   false,
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

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			controller := New(logger, nil, mockAuthorUseCase)

			if tc.shouldCall {
				mockAuthorUseCase.EXPECT().
					GetAuthorInfo(ctx, logger, tc.req.GetId()).
					Return(tc.expectedResp, tc.expectedErr)
			}

			response, err := controller.GetAuthorInfo(ctx, tc.req)
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
		books       []dto.Book
		expectedErr error
		errStatus   codes.Code
		shouldCall  bool
	}{
		{
			name: "TestGetAuthorBooks",
			req: &library2.GetAuthorBooksRequest{
				AuthorId: "550e8400-e29b-41d4-a716-446655440000",
			},
			books: []dto.Book{
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
			shouldCall:  true,
		},
		{
			name: "InvalidAuthorID",
			req: &library2.GetAuthorBooksRequest{
				AuthorId: "invalid-id",
			},
			books:       nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
			shouldCall:  false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			logger := zap.NewNop()

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			mockStream := mocks.NewMockLibrary_GetAuthorBooksServer(ctrl)
			controller := New(logger, nil, mockAuthorUseCase)

			mockStream.EXPECT().Context().Return(ctx).AnyTimes()

			if tc.shouldCall {
				bookCh := make(chan dto.Book, len(tc.books))
				errCh := make(chan error, 1)

				mockAuthorUseCase.EXPECT().
					StreamBooksForAuthor(ctx, logger, tc.req.GetAuthorId()).
					Return((<-chan dto.Book)(bookCh), (<-chan error)(errCh))

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
			}

			err := controller.GetAuthorBooks(tc.req, mockStream)

			if tc.errStatus != codes.OK {
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

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name        string
		req         *library2.UpdateBookRequest
		resp        *library2.UpdateBookResponse
		expectedErr error
		errStatus   codes.Code
		shouldCall  bool
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
			shouldCall:  true,
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
			shouldCall:  true,
		},
		{
			name: "InvalidBookID",
			req: &library2.UpdateBookRequest{
				Id:        "invalid-id",
				Name:      "Author 1",
				AuthorIds: []string{"123e4567-e89b-12d3-a456-426614174000"},
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
			shouldCall:  false,
		},
		{
			name: "TooManyAuthors",
			req: &library2.UpdateBookRequest{
				Id:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Valid Name",
				AuthorIds: InvalidAuthorsIDs,
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
			shouldCall:  false,
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

			mockBooksUseCase := mocks.NewMockBooksUseCase(ctrl)
			controller := New(logger, mockBooksUseCase, nil)

			if tc.shouldCall {
				mockBooksUseCase.EXPECT().
					UpdateBook(ctx, logger, tc.req.GetId(), tc.req.GetName(), tc.req.GetAuthorIds()).
					Return(tc.expectedErr)
			}

			response, err := controller.UpdateBook(ctx, tc.req)
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
		shouldCall  bool
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
			shouldCall:  true,
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
			shouldCall:  true,
		},
		{
			name: "InvalidAuthorID",
			req: &library2.ChangeAuthorInfoRequest{
				Id:   "invalid-id",
				Name: "newName",
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
			shouldCall:  false,
		},
		{
			name: "EmptyName",
			req: &library2.ChangeAuthorInfoRequest{
				Id:   "550e8400-e29b-41d4-a716-446655440000",
				Name: "",
			},
			resp:        nil,
			expectedErr: nil,
			errStatus:   codes.InvalidArgument,
			shouldCall:  false,
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

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			controller := New(logger, nil, mockAuthorUseCase)

			if tc.shouldCall {
				mockAuthorUseCase.EXPECT().
					ChangeAuthorInfo(ctx, logger, tc.req.GetId(), tc.req.GetName()).
					Return(tc.expectedErr)
			}

			response, err := controller.ChangeAuthorInfo(ctx, tc.req)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.errStatus, s.Code())
			require.Equal(t, tc.resp, response)
		})
	}
}

func TestConvertErr(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		input    error
		expected codes.Code
	}{
		{
			name:     "ErrAuthorNotFound",
			input:    entity.ErrAuthorNotFound,
			expected: codes.NotFound,
		},
		{
			name:     "ErrBookNotFound",
			input:    entity.ErrBookNotFound,
			expected: codes.NotFound,
		},
		{
			name:     "ErrBookAlreadyExists",
			input:    entity.ErrBookAlreadyExists,
			expected: codes.AlreadyExists,
		},
		{
			name:     "ErrAuthorAlreadyExists",
			input:    entity.ErrAuthorAlreadyExists,
			expected: codes.AlreadyExists,
		},
		{
			name:     "InternalError",
			input:    errors.New("some internal error"),
			expected: codes.Internal,
		},
		{
			name:     "NilError",
			input:    nil,
			expected: codes.OK,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := &implementation{}
			err := i.convertErr(tc.input)

			if tc.input == nil {
				require.NoError(t, err)
				return
			}

			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.expected, s.Code())
			require.Equal(t, tc.input.Error(), s.Message())
		})
	}
}
