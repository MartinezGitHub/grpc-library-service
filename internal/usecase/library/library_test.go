package library

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/dto"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/mocks"
	"github.com/project/library/internal/usecase/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name           string
		expectedAuthor entity.Author
		expectedErr    error
		authorName     string
		wantErr        error
		wantResponse   *library.RegisterAuthorResponse
	}{
		{
			name: "TestRegisterAuthor",
			expectedAuthor: entity.Author{
				ID:   "123",
				Name: "Author 1",
			},
			expectedErr:  nil,
			authorName:   "Author 1",
			wantErr:      nil,
			wantResponse: &library.RegisterAuthorResponse{Id: "123"},
		},
		{
			name:           "ErrAuthorAlreadyExists",
			expectedAuthor: entity.Author{},
			expectedErr:    entity.ErrAuthorAlreadyExists,
			authorName:     "Author 1",
			wantErr:        entity.ErrAuthorAlreadyExists,
			wantResponse:   nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
			mockOutbox := mocks.NewMockOutboxRepository(ctrl)
			mockTransactor := mocks.NewMockTransactor(ctrl)

			lib := New(nil, mockAuthor, nil, mockOutbox, mockTransactor)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockTransactor.EXPECT().
				WithTx(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})

			mockAuthor.EXPECT().
				CreateAuthor(ctx, logger, gomock.Any()).
				Return(tc.expectedAuthor, tc.expectedErr)

			if tc.expectedErr == nil {
				serialized, _ := json.Marshal(tc.expectedAuthor)
				mockOutbox.EXPECT().
					SendMessage(ctx, "author_"+tc.expectedAuthor.ID, repository.OutboxKindAuthor, serialized).
					Return(nil)
			}

			resp, err := lib.RegisterAuthor(ctx, logger, tc.authorName)
			require.ErrorIs(t, err, tc.wantErr)
			require.Equal(t, tc.wantResponse, resp)
		})
	}
}

func TestRegisterBook(t *testing.T) {
	t.Parallel()
	now := time.Now()
	testcases := []struct {
		name         string
		bookName     string
		authorIDs    []string
		expectedBook entity.Book
		expectedErr  error
		wantErr      error
		wantResponse *library.AddBookResponse
	}{
		{
			name:      "TestRegisterBook",
			bookName:  "Book 1",
			authorIDs: []string{"1", "2"},
			expectedBook: entity.Book{
				ID:        "1",
				Name:      "Book 1",
				AuthorIDs: []string{"1", "2"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedErr: nil,
			wantErr:     nil,
			wantResponse: &library.AddBookResponse{
				Book: &library.Book{
					Id:        "1",
					Name:      "Book 1",
					AuthorId:  []string{"1", "2"},
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
			},
		},
		{
			name:      "ErrAuthorNotFound",
			bookName:  "Book 2",
			authorIDs: []string{"1", "2"},
			expectedBook: entity.Book{
				ID:        "2",
				Name:      "Book 2",
				AuthorIDs: []string{"1", "2"},
			},
			expectedErr:  entity.ErrAuthorNotFound,
			wantErr:      entity.ErrAuthorNotFound,
			wantResponse: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBook := mocks.NewMockBooksRepository(ctrl)
			mockOutbox := mocks.NewMockOutboxRepository(ctrl)
			mockTransactor := mocks.NewMockTransactor(ctrl)

			lib := New(nil, nil, mockBook, mockOutbox, mockTransactor)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockTransactor.EXPECT().
				WithTx(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				})

			mockBook.EXPECT().
				CreateBook(ctx, logger, gomock.Any()).
				Return(tc.expectedBook, tc.expectedErr)

			if tc.expectedErr == nil {
				serialized, _ := json.Marshal(tc.expectedBook)
				mockOutbox.EXPECT().
					SendMessage(ctx, "book_"+tc.expectedBook.ID, repository.OutboxKindBook, serialized).
					Return(nil)
			}

			resp, err := lib.RegisterBook(ctx, logger, tc.bookName, tc.authorIDs)
			require.ErrorIs(t, err, tc.wantErr)
			if tc.wantResponse != nil {
				require.Equal(t, tc.wantResponse.GetBook().GetId(), resp.GetBook().GetId())
				require.Equal(t, tc.wantResponse.GetBook().GetName(), resp.GetBook().GetName())
				require.ElementsMatch(t, tc.wantResponse.GetBook().GetAuthorId(), resp.GetBook().GetAuthorId())
			} else {
				require.Nil(t, resp)
			}
		})
	}
}

func TestStreamBooksForAuthor(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name          string
		authorID      string
		expectedBooks []entity.Book
		expectedErr   error
		wantBooks     []dto.Book
		wantErr       error
	}{
		{
			name:     "TestGetBooksForAuthor one book",
			authorID: "1",
			expectedBooks: []entity.Book{
				{
					ID:        "1",
					Name:      "Book 1",
					AuthorIDs: []string{"1"},
				},
			},
			expectedErr: nil,
			wantBooks: []dto.Book{
				{
					ID:        "1",
					Name:      "Book 1",
					AuthorIDs: []string{"1"},
				},
			},
			wantErr: nil,
		},
		{
			name:     "multiple books",
			authorID: "1",
			expectedBooks: []entity.Book{
				{ID: "1", Name: "Book 1", AuthorIDs: []string{"1"}},
				{ID: "2", Name: "Book 2", AuthorIDs: []string{"1", "2"}},
			},
			wantBooks: []dto.Book{
				{ID: "1", Name: "Book 1", AuthorIDs: []string{"1"}},
				{ID: "2", Name: "Book 2", AuthorIDs: []string{"1", "2"}},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBook := mocks.NewMockBooksRepository(ctrl)
			lib := New(nil, nil, mockBook, nil, nil)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			booksCh := make(chan entity.Book, len(tc.expectedBooks))
			errCh := make(chan error, 1)

			if tc.expectedErr != nil {
				errCh <- tc.expectedErr
			} else {
				for _, book := range tc.expectedBooks {
					booksCh <- book
				}
			}
			close(booksCh)
			close(errCh)

			mockBook.EXPECT().
				StreamBooksByAuthorID(gomock.Any(), gomock.Any(), tc.authorID).
				Return(booksCh, errCh).
				Times(1)

			resultBooksCh, resultErrCh := lib.StreamBooksForAuthor(ctx, logger, tc.authorID)

			var receivedBooks []dto.Book
			var receivedErr error

			for book := range resultBooksCh {
				receivedBooks = append(receivedBooks, book)
			}
			require.ElementsMatch(t, tc.wantBooks, receivedBooks)
			receivedErr = <-resultErrCh
			require.Equal(t, tc.wantErr, receivedErr)
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name           string
		authorID       string
		expectedAuthor entity.Author
		expectedErr    error
		wantResponse   *library.GetAuthorInfoResponse
		wantErr        error
	}{
		{
			name:        "TestGetAuthorInfo",
			authorID:    "1",
			expectedErr: nil,
			expectedAuthor: entity.Author{
				ID:   "1",
				Name: "Author 1",
			},
			wantResponse: &library.GetAuthorInfoResponse{
				Id:   "1",
				Name: "Author 1",
			},
			wantErr: nil,
		},
		{
			name:           "ErrAuthorNotFound",
			authorID:       "2",
			expectedErr:    entity.ErrAuthorNotFound,
			expectedAuthor: entity.Author{},
			wantResponse:   nil,
			wantErr:        entity.ErrAuthorNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
			lib := New(nil, mockAuthor, nil, nil, nil)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockAuthor.EXPECT().
				GetAuthor(ctx, logger, tc.authorID).
				Return(tc.expectedAuthor, tc.expectedErr)

			resp, err := lib.GetAuthorInfo(ctx, logger, tc.authorID)
			require.ErrorIs(t, err, tc.wantErr)
			require.Equal(t, tc.wantResponse, resp)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()
	now := time.Now()
	testcases := []struct {
		name         string
		bookID       string
		expectedBook entity.Book
		expectedErr  error
		wantResponse *library.GetBookInfoResponse
		wantErr      error
	}{
		{
			name:   "TestGetBook",
			bookID: "1",
			expectedBook: entity.Book{
				ID:        "1",
				Name:      "Book 1",
				AuthorIDs: []string{"1", "2"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedErr: nil,
			wantResponse: &library.GetBookInfoResponse{
				Book: &library.Book{
					Id:        "1",
					Name:      "Book 1",
					AuthorId:  []string{"1", "2"},
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				},
			},
			wantErr: nil,
		},
		{
			name:         "ErrBookNotFound",
			bookID:       "2",
			expectedBook: entity.Book{},
			expectedErr:  entity.ErrBookNotFound,
			wantResponse: nil,
			wantErr:      entity.ErrBookNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBook := mocks.NewMockBooksRepository(ctrl)
			lib := New(nil, nil, mockBook, nil, nil)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockBook.EXPECT().
				GetBook(ctx, logger, tc.bookID).
				Return(tc.expectedBook, tc.expectedErr)

			resp, err := lib.GetBook(ctx, logger, tc.bookID)
			require.ErrorIs(t, err, tc.wantErr)
			require.Equal(t, tc.wantResponse, resp)
		})
	}
}

func TestChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name        string
		authorID    string
		newName     string
		expectedErr error
		wantErr     error
	}{
		{
			name:        "Success",
			authorID:    "1",
			newName:     "New Author Name",
			expectedErr: nil,
			wantErr:     nil,
		},
		{
			name:        "AuthorNotFound",
			authorID:    "2",
			newName:     "Non-existent Author",
			expectedErr: entity.ErrAuthorNotFound,
			wantErr:     entity.ErrAuthorNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
			lib := New(nil, mockAuthor, nil, nil, nil)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockAuthor.EXPECT().
				ChangeAuthorForID(ctx, logger, tc.authorID, tc.newName).
				Return(tc.expectedErr)

			err := lib.ChangeAuthorInfo(ctx, logger, tc.authorID, tc.newName)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name        string
		bookID      string
		bookName    string
		authorIDs   []string
		expectedErr error
		wantErr     error
	}{
		{
			name:        "Success",
			bookID:      "1",
			bookName:    "Updated Book Name",
			authorIDs:   []string{"1", "2"},
			expectedErr: nil,
			wantErr:     nil,
		},
		{
			name:        "BookNotFound",
			bookID:      "999",
			bookName:    "Non-existent Book",
			authorIDs:   []string{"1"},
			expectedErr: entity.ErrBookNotFound,
			wantErr:     entity.ErrBookNotFound,
		},
		{
			name:        "AuthorNotFound",
			bookID:      "1",
			bookName:    "Book With Bad Author",
			authorIDs:   []string{"999"},
			expectedErr: entity.ErrAuthorNotFound,
			wantErr:     entity.ErrAuthorNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBook := mocks.NewMockBooksRepository(ctrl)
			lib := New(nil, nil, mockBook, nil, nil)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockBook.EXPECT().
				UpdateBookByID(ctx, logger, tc.bookID, tc.bookName, tc.authorIDs).
				Return(tc.expectedErr)

			err := lib.UpdateBook(ctx, logger, tc.bookID, tc.bookName, tc.authorIDs)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestGetAuthorByID(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name           string
		authorID       string
		expectedAuthor entity.Author
		expectedErr    error
		wantAuthor     entity.Author
		wantErr        error
	}{
		{
			name:     "Success",
			authorID: "1",
			expectedAuthor: entity.Author{
				ID:   "1",
				Name: "Author 1",
			},
			expectedErr: nil,
			wantAuthor: entity.Author{
				ID:   "1",
				Name: "Author 1",
			},
			wantErr: nil,
		},
		{
			name:           "AuthorNotFound",
			authorID:       "999",
			expectedAuthor: entity.Author{},
			expectedErr:    entity.ErrAuthorNotFound,
			wantAuthor:     entity.Author{},
			wantErr:        entity.ErrAuthorNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
			lib := New(nil, mockAuthor, nil, nil, nil)
			ctx := context.Background()
			logger, _ := zap.NewProduction()

			mockAuthor.EXPECT().
				GetAuthor(ctx, logger, tc.authorID).
				Return(tc.expectedAuthor, tc.expectedErr)

			author, err := lib.GetAuthorByID(ctx, logger, tc.authorID)
			require.ErrorIs(t, err, tc.wantErr)
			require.Equal(t, tc.wantAuthor, author)
		})
	}
}

func TestRegisterBookOutboxErrorOnly(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBook := mocks.NewMockBooksRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)
	mockTransactor := mocks.NewMockTransactor(ctrl)

	mockTransactor.EXPECT().WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		})

	mockBook.EXPECT().CreateBook(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Book{
			ID:        "123",
			Name:      "Test Book",
			AuthorIDs: []string{"1"},
		}, nil)

	outboxErr := errors.New("outbox send error")
	mockOutbox.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(outboxErr)

	lib := New(nil, nil, mockBook, mockOutbox, mockTransactor)
	ctx := context.Background()
	logger, _ := zap.NewProduction()

	resp, err := lib.RegisterBook(ctx, logger, "Test Book", []string{"1"})

	require.Error(t, err)
	require.Equal(t, outboxErr, err)
	require.Nil(t, resp)
}

func TestRegisterAuthorOutboxErrorOnly(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthor := mocks.NewMockAuthorRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)
	mockTransactor := mocks.NewMockTransactor(ctrl)

	mockTransactor.EXPECT().WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		})

	mockAuthor.EXPECT().CreateAuthor(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Author{
			ID:   "123",
			Name: "Test Author",
		}, nil)

	outboxErr := errors.New("outbox send error")
	mockOutbox.EXPECT().SendMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(outboxErr)

	lib := New(nil, mockAuthor, nil, mockOutbox, mockTransactor)
	ctx := context.Background()
	logger, _ := zap.NewProduction()

	resp, err := lib.RegisterAuthor(ctx, logger, "Test Author")

	require.Error(t, err)
	require.Equal(t, outboxErr, err)
	require.Nil(t, resp)
}
