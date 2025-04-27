package library

//import (
//	"context"
//	"testing"
//
//	"github.com/golang/mock/gomock"
//	"github.com/project/library/internal/entity"
//	"github.com/project/library/internal/mocks"
//	log "github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/require"
//	"go.uber.org/zap"
//)
//
//func TestRegisterAuthor(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name           string
//		expectedAuthor entity.Author
//		expectedErr    error
//		authorName     string
//		wantErr        error
//		wantID         string
//	}{
//		{
//			name: "TestRegisterAuthor",
//			expectedAuthor: entity.Author{
//				ID:   "123",
//				Name: "Author 1",
//			},
//			expectedErr: nil,
//			authorName:  "Author 1",
//			wantErr:     nil,
//			wantID:      "123",
//		},
//		{
//			name:           "ErrAuthorAlreadyExists",
//			expectedAuthor: entity.Author{},
//			expectedErr:    entity.ErrAuthorAlreadyExists,
//			authorName:     "Author 1",
//			wantErr:        entity.ErrAuthorAlreadyExists,
//			wantID:         "",
//		},
//	}
//	ctrl := gomock.NewController(t)
//	t.Cleanup(func() { ctrl.Finish() })
//	mockRepo := mocks.NewMockAuthorRepository(ctrl)
//	lib := New(nil, mockRepo, nil)
//	ctx := context.Background()
//	var logger *zap.Logger
//
//	logger, err := zap.NewProduction()
//	if err != nil {
//		ctrl.Finish()
//		log.Fatalf("can not initialize logger: %s", err)
//	}
//
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			mockRepo.EXPECT().CreateAuthor(ctx, logger, gomock.Any()).Return(tc.expectedAuthor, tc.expectedErr)
//			id, err := lib.RegisterAuthor(ctx, logger, "Author 1")
//			require.ErrorIs(t, err, tc.wantErr)
//			require.Equal(t, tc.expectedAuthor.ID, id)
//		})
//	}
//}
//
//func TestChangeAuthorInfo(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name             string
//		authorID         string
//		existedAuthorIDs []string
//		expectedErr      error
//		newAuthorName    string
//		wantErr          error
//	}{
//		{
//			name:             "NoErr",
//			authorID:         "123",
//			existedAuthorIDs: []string{"123"},
//			expectedErr:      nil,
//			newAuthorName:    "Author 1",
//			wantErr:          nil,
//		},
//		{
//			name:             "ErrAuthorAlreadyExists",
//			authorID:         "123",
//			existedAuthorIDs: []string{},
//			expectedErr:      entity.ErrAuthorNotFound,
//			newAuthorName:    "Author 2",
//			wantErr:          entity.ErrAuthorNotFound,
//		},
//	}
//
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//			mockRepo := mocks.NewMockAuthorRepository(ctrl)
//			lib := New(nil, mockRepo, nil)
//			ctx := context.Background()
//
//			logger, err := zap.NewProduction()
//			if err != nil {
//				log.Fatalf("can not initialize logger: %s", err)
//			}
//			mockRepo.EXPECT().ChangeAuthorForID(ctx, logger, gomock.Any(), gomock.Any()).Return(tc.expectedErr)
//
//			err = lib.ChangeAuthorInfo(ctx, logger, tc.authorID, tc.newAuthorName)
//			require.ErrorIs(t, err, tc.wantErr)
//		})
//	}
//}
//
//func TestGetBooksForAuthor(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name          string
//		authorID      string
//		expectedBooks []entity.Book
//		expectedErr   error
//		wantBooks     []entity.Book
//		wantErr       error
//	}{
//		{
//			name:     "TestGetBooksForAuthor one book",
//			authorID: "1",
//			expectedBooks: []entity.Book{
//				{
//					ID:        "1",
//					Name:      "Book 1",
//					AuthorIDs: []string{"1"},
//				},
//			},
//			expectedErr: nil,
//			wantBooks: []entity.Book{
//				{
//					ID:        "1",
//					Name:      "Book 1",
//					AuthorIDs: []string{"1"},
//				},
//			},
//			wantErr: nil,
//		},
//		{
//			name:     "multiple books",
//			authorID: "1",
//			expectedBooks: []entity.Book{
//				{ID: "1", Name: "Book 1", AuthorIDs: []string{"1"}},
//				{ID: "2", Name: "Book 2", AuthorIDs: []string{"1", "2"}},
//			},
//			wantBooks: []entity.Book{
//				{ID: "1", Name: "Book 1", AuthorIDs: []string{"1"}},
//				{ID: "2", Name: "Book 2", AuthorIDs: []string{"1", "2"}},
//			},
//			wantErr: nil,
//		},
//	}
//
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//
//			mockBook := mocks.NewMockBooksRepository(ctrl)
//			lib := New(nil, nil, mockBook)
//			ctx := context.Background()
//
//			logger, err := zap.NewProduction()
//			if err != nil {
//				log.Fatalf("can not initialize logger: %s", err)
//			}
//
//			booksCh := make(chan entity.Book, 10)
//			errCh := make(chan error, 1)
//
//			if tc.expectedErr != nil {
//				errCh <- tc.expectedErr
//			} else {
//				for _, book := range tc.expectedBooks {
//					booksCh <- book
//				}
//			}
//			close(booksCh)
//			close(errCh)
//
//			mockBook.EXPECT().
//				StreamBooksByAuthorID(gomock.Any(), gomock.Any(), tc.authorID).
//				Return(booksCh, errCh).
//				Times(1)
//
//			resultBooksCh, resultErrCh := lib.StreamBooksForAuthor(ctx, logger, tc.authorID)
//
//			var receivedBooks []entity.Book
//			var receivedErr error
//
//			for book := range resultBooksCh {
//				receivedBooks = append(receivedBooks, book)
//			}
//			require.ElementsMatch(t, tc.wantBooks, receivedBooks)
//			receivedErr = <-resultErrCh
//			require.Equal(t, tc.wantErr, receivedErr)
//		})
//	}
//}
//
//func TestGetAuthorByID(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name           string
//		authorID       string
//		expectedErr    error
//		expectedAuthor entity.Author
//		wantAuthor     entity.Author
//		wantErr        error
//	}{
//		{
//			name:        "TestGetAuthorByID",
//			authorID:    "1",
//			expectedErr: nil,
//			expectedAuthor: entity.Author{
//				ID:   "1",
//				Name: "Author 1",
//			},
//			wantAuthor: entity.Author{
//				ID:   "1",
//				Name: "Author 1",
//			},
//			wantErr: nil,
//		},
//		{
//			name:           "ErrAuthorNotFound",
//			authorID:       "2",
//			expectedErr:    entity.ErrAuthorNotFound,
//			expectedAuthor: entity.Author{},
//			wantErr:        entity.ErrAuthorNotFound,
//			wantAuthor:     entity.Author{},
//		},
//	}
//	ctrl := gomock.NewController(t)
//	t.Cleanup(func() { ctrl.Finish() })
//	mockRepo := mocks.NewMockAuthorRepository(ctrl)
//	lib := New(nil, mockRepo, nil)
//	ctx := context.Background()
//	logger, err := zap.NewProduction()
//	if err != nil {
//		ctrl.Finish()
//		log.Fatalf("can not initialize logger: %s", err)
//	}
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			mockRepo.EXPECT().GetAuthor(ctx, logger, gomock.Any()).Return(tc.expectedAuthor, tc.expectedErr)
//			author, err := lib.GetAuthorByID(ctx, logger, tc.authorID)
//			require.ErrorIs(t, err, tc.wantErr)
//			require.Equal(t, tc.wantAuthor, author)
//		})
//	}
//}
//
//func TestGetAuthorInfo(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name           string
//		authorID       string
//		expectedErr    error
//		expectedAuthor entity.Author
//		wantName       string
//		wantErr        error
//	}{
//		{
//			name:        "TestGetAuthorInfo",
//			authorID:    "1",
//			expectedErr: nil,
//			expectedAuthor: entity.Author{
//				ID:   "1",
//				Name: "Author 1",
//			},
//			wantName: "Author 1",
//			wantErr:  nil,
//		},
//		{
//			name:           "ErrAuthorNotFound",
//			authorID:       "2",
//			expectedErr:    entity.ErrAuthorNotFound,
//			expectedAuthor: entity.Author{},
//			wantName:       "",
//			wantErr:        entity.ErrAuthorNotFound,
//		},
//	}
//	ctrl := gomock.NewController(t)
//	t.Cleanup(func() { ctrl.Finish() })
//	mockRepo := mocks.NewMockAuthorRepository(ctrl)
//	lib := New(nil, mockRepo, nil)
//	ctx := context.Background()
//	logger, err := zap.NewProduction()
//	if err != nil {
//		log.Fatalf("can not initialize logger: %s", err)
//	}
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			mockRepo.EXPECT().GetAuthor(ctx, logger, gomock.Any()).Return(tc.expectedAuthor, tc.expectedErr)
//			name, err := lib.GetAuthorInfo(ctx, logger, tc.authorID)
//			require.ErrorIs(t, err, tc.wantErr)
//			require.Equal(t, tc.wantName, name)
//		})
//	}
//}
//
//func TestRegisterBook(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name           string
//		bookName       string
//		authorIDs      []string
//		existAuthorIDs []string
//		expectedBook   entity.Book
//		expectedErr    error
//		wantErr        error
//		wantBook       entity.Book
//	}{
//		{
//			name:           "TestRegisterBook",
//			bookName:       "Book 1",
//			authorIDs:      []string{"1", "2"},
//			existAuthorIDs: []string{"1", "2"},
//			expectedBook: entity.Book{
//				ID:        "1",
//				Name:      "Book 1",
//				AuthorIDs: []string{"1", "2"},
//			},
//			expectedErr: nil,
//			wantErr:     nil,
//			wantBook: entity.Book{
//				ID:        "1",
//				Name:      "Book 1",
//				AuthorIDs: []string{"1", "2"},
//			},
//		},
//		{
//			name:           "ErrAuthorNotFound",
//			bookName:       "Book 2",
//			authorIDs:      []string{"1", "2"},
//			existAuthorIDs: []string{"1"},
//			expectedBook:   entity.Book{},
//			expectedErr:    entity.ErrAuthorNotFound,
//			wantErr:        entity.ErrAuthorNotFound,
//			wantBook:       entity.Book{},
//		},
//		{
//			name:           "ErrBookAlreadyExists",
//			bookName:       "Book 3",
//			authorIDs:      []string{"3"},
//			existAuthorIDs: []string{"3"},
//			expectedBook:   entity.Book{},
//			expectedErr:    entity.ErrBookAlreadyExists,
//			wantBook:       entity.Book{},
//			wantErr:        entity.ErrBookAlreadyExists,
//		},
//	}
//
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//			ctx := context.Background()
//			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
//			mockBook := mocks.NewMockBooksRepository(ctrl)
//			lib := New(nil, mockAuthor, mockBook)
//
//			logger, err := zap.NewProduction()
//			if err != nil {
//				log.Fatalf("can not initialize logger: %s", err)
//			}
//
//			mockBook.EXPECT().CreateBook(ctx, logger, gomock.Any()).Return(tc.expectedBook, tc.expectedErr)
//
//			book, err := lib.RegisterBook(ctx, logger, tc.bookName, tc.authorIDs)
//			require.ErrorIs(t, err, tc.wantErr)
//			require.Equal(t, tc.wantBook, book)
//		})
//	}
//}
//
//func TestGetBook(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name         string
//		bookID       string
//		expectedBook entity.Book
//		expectedErr  error
//		wantBook     entity.Book
//		wantErr      error
//	}{
//		{
//			name:   "TestGetBook",
//			bookID: "1",
//			expectedBook: entity.Book{
//				ID:        "1",
//				Name:      "Book 1",
//				AuthorIDs: []string{"1", "2"},
//			},
//			expectedErr: nil,
//			wantBook: entity.Book{
//				ID:        "1",
//				Name:      "Book 1",
//				AuthorIDs: []string{"1", "2"},
//			},
//			wantErr: nil,
//		},
//		{
//			name:         "ErrBookNotFound",
//			bookID:       "2",
//			expectedBook: entity.Book{},
//			expectedErr:  entity.ErrBookNotFound,
//			wantBook:     entity.Book{},
//			wantErr:      entity.ErrBookNotFound,
//		},
//	}
//
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//			ctx := context.Background()
//			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
//			mockBook := mocks.NewMockBooksRepository(ctrl)
//			lib := New(nil, mockAuthor, mockBook)
//			logger, err := zap.NewProduction()
//			if err != nil {
//				log.Fatalf("can not initialize logger: %s", err)
//			}
//
//			mockBook.EXPECT().GetBook(ctx, logger, gomock.Any()).Return(tc.expectedBook, tc.expectedErr)
//
//			book, err := lib.GetBook(ctx, logger, tc.bookID)
//			require.ErrorIs(t, err, tc.wantErr)
//			require.Equal(t, tc.wantBook, book)
//		})
//	}
//}
//
//func TestUpdateBook(t *testing.T) {
//	t.Parallel()
//	testcases := []struct {
//		name         string
//		bookID       string
//		bookName     string
//		authorIDs    []string
//		expectedBook entity.Book
//		expectedErr  error
//		wantErr      error
//		wantBook     entity.Book
//	}{
//		{
//			name:      "TestUpdateBook",
//			bookID:    "1",
//			bookName:  "Book 11",
//			authorIDs: []string{"1", "2"},
//			expectedBook: entity.Book{
//				ID:        "1",
//				Name:      "Book 11",
//				AuthorIDs: []string{"1", "2"},
//			},
//			expectedErr: nil,
//			wantErr:     nil,
//			wantBook: entity.Book{
//				ID:        "1",
//				Name:      "Book 11",
//				AuthorIDs: []string{"1", "2"},
//			},
//		},
//		{
//			name:      "ErrAuthorNotFound",
//			bookID:    "2",
//			bookName:  "Book 2",
//			authorIDs: []string{"1", "2"},
//
//			expectedBook: entity.Book{},
//			expectedErr:  entity.ErrAuthorNotFound,
//			wantErr:      entity.ErrAuthorNotFound,
//			wantBook:     entity.Book{},
//		},
//		{
//			name:         "ErrBookNotFound",
//			bookID:       "3",
//			bookName:     "Book 3",
//			authorIDs:    []string{},
//			expectedBook: entity.Book{},
//			expectedErr:  entity.ErrBookNotFound,
//			wantErr:      entity.ErrBookNotFound,
//			wantBook:     entity.Book{},
//		},
//	}
//
//	for _, tc := range testcases {
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//			ctx := context.Background()
//			mockAuthor := mocks.NewMockAuthorRepository(ctrl)
//			mockBook := mocks.NewMockBooksRepository(ctrl)
//			lib := New(nil, mockAuthor, mockBook)
//			logger, err := zap.NewProduction()
//			if err != nil {
//				log.Fatalf("can not initialize logger: %s", err)
//			}
//
//			mockBook.EXPECT().UpdateBookByID(ctx, logger, gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.expectedErr)
//
//			err = lib.UpdateBook(ctx, logger, tc.bookID, tc.bookName, tc.authorIDs)
//			require.ErrorIs(t, err, tc.wantErr)
//			if err == nil {
//				mockBook.EXPECT().GetBook(ctx, logger, tc.bookID).Return(tc.expectedBook, tc.expectedErr)
//				book, err2 := mockBook.GetBook(ctx, logger, tc.bookID)
//				require.Equal(t, tc.expectedBook, book)
//				require.NoError(t, err2)
//			}
//		})
//	}
//}
