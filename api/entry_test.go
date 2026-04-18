package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/KamisAyaka/simplebank/db/mock"
	db "github.com/KamisAyaka/simplebank/db/sqlc"
	"github.com/KamisAyaka/simplebank/token"
	"github.com/KamisAyaka/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetEntryAPI(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)
	entry := randomEntry()
	entry.AccountID = account.ID

	testCases := []struct {
		name      string
		entryID   int64
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		// gomock 期望定义：接口调用签名和返回结果。
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(entry, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(entry.AccountID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name:    "UnauthorizedUser",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "another-user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(entry, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(entry.AccountID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:    "NoAuthorization",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:    "NotFound",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, sql.ErrConnDone)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:    "InvalidID",
			entryID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// 每个子测试独立一个 Controller，避免期望串扰。
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries/%d", tc.entryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			// 真正执行 HTTP 请求并走到 handler。
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListEntriesAPI(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)
	entries := []db.Entry{randomEntry(), randomEntry(), randomEntry()}
	for i := range entries {
		entries[i].AccountID = account.ID
	}

	testCases := []struct {
		name          string
		accountID     int64
		pageID        int32
		pageSize      int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			pageID:    1,
			pageSize:  10,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: account.ID,
					Limit:     10,
					Offset:    0,
				}
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
				store.EXPECT().ListEntries(gomock.Any(), gomock.Eq(arg)).Times(1).Return(entries, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntries(t, recorder.Body, entries)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			pageID:    1,
			pageSize:  10,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "another-user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "InvalidPageSize",
			accountID: account.ID,
			pageID:    1,
			pageSize:  0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			pageID:    1,
			pageSize:  10,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			pageID:    1,
			pageSize:  10,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: account.ID,
					Limit:     10,
					Offset:    0,
				}
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
				store.EXPECT().ListEntries(gomock.Any(), gomock.Eq(arg)).Times(1).Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries?account_id=%d&page_id=%d&page_size=%d", tc.accountID, tc.pageID, tc.pageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomEntry() db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 10000),
		AccountID: util.RandomInt(1, 10000),
		Amount:    util.RandomEntryAmount(),
		CreatedAt: time.Now(),
	}
}

func requireBodyMatchEntry(t *testing.T, body *bytes.Buffer, entry db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntry db.Entry
	err = json.Unmarshal(data, &gotEntry)
	require.NoError(t, err)
	require.Equal(t, entry.ID, gotEntry.ID)
	require.Equal(t, entry.AccountID, gotEntry.AccountID)
	require.Equal(t, entry.Amount, gotEntry.Amount)
	require.WithinDuration(t, entry.CreatedAt, gotEntry.CreatedAt, time.Second)
}

func requireBodyMatchEntries(t *testing.T, body *bytes.Buffer, entries []db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntries []db.Entry
	err = json.Unmarshal(data, &gotEntries)
	require.NoError(t, err)
	require.Len(t, gotEntries, len(entries))
	for i := range entries {
		require.Equal(t, entries[i].ID, gotEntries[i].ID)
		require.Equal(t, entries[i].AccountID, gotEntries[i].AccountID)
		require.Equal(t, entries[i].Amount, gotEntries[i].Amount)
		require.WithinDuration(t, entries[i].CreatedAt, gotEntries[i].CreatedAt, time.Second)
	}
}
