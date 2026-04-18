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
	"github.com/KamisAyaka/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferAPI(t *testing.T) {
	fromAccount := randomAccountWithCurrency("USD")
	toAccount := randomAccountWithCurrency("USD")
	amount := int64(10)
	txResult := randomTransferTxResult(fromAccount, toAccount, amount)

	testCases := []struct {
		name string
		body gin.H
		// buildStubs 描述这个场景下 store 应该被如何调用。
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amount,
				}
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(txResult, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferTxResult(t, recorder.Body, txResult)
			},
		},
		{
			name: "InvalidRequest",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   fromAccount.ID,
				"amount":          -10,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "CurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				mismatch := randomAccountWithCurrency("EUR")
				mismatch.ID = fromAccount.ID
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(mismatch, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amount,
				}
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// Controller + MockStore 是 gomock 的标准初始化写法。
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(data))
			require.NoError(t, err)

			// 触发 handler 后，gomock 会在结束时校验是否按期望调用。
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetTransferAPI(t *testing.T) {
	transfer := randomTransfer()

	testCases := []struct {
		name          string
		transferID    int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(transfer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},
		{
			name:       "NotFound",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(db.Transfer{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			transferID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(db.Transfer{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/transfers/%d", tc.transferID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListTransfersAPI(t *testing.T) {
	transfers := []db.Transfer{randomTransfer(), randomTransfer(), randomTransfer()}

	testCases := []struct {
		name          string
		accountID     int64
		pageID        int32
		pageSize      int32
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: 1,
			pageID:    1,
			pageSize:  10,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListTransfersParams{
					FromAccountID: 1,
					Limit:         10,
					Offset:        0,
				}
				store.EXPECT().ListTransfers(gomock.Any(), gomock.Eq(arg)).Times(1).Return(transfers, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfers(t, recorder.Body, transfers)
			},
		},
		{
			name:      "InvalidPageSize",
			accountID: 1,
			pageID:    1,
			pageSize:  0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfers(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: 1,
			pageID:    1,
			pageSize:  10,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListTransfersParams{
					FromAccountID: 1,
					Limit:         10,
					Offset:        0,
				}
				store.EXPECT().ListTransfers(gomock.Any(), gomock.Eq(arg)).Times(1).Return(nil, sql.ErrConnDone)
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

			url := fmt.Sprintf("/transfers?account_id=%d&page_id=%d&page_size=%d", tc.accountID, tc.pageID, tc.pageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomTransfer() db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 10000),
		FromAccountID: util.RandomInt(1, 10000),
		ToAccountID:   util.RandomInt(1, 10000),
		Amount:        util.RandomPositiveMoney(),
		CreatedAt:     time.Now(),
	}
}

func randomAccountWithCurrency(currency string) db.Account {
	account := randomAccount()
	account.Currency = currency
	return account
}

func randomTransferTxResult(fromAccount db.Account, toAccount db.Account, amount int64) db.TransferTxResult {
	transfer := randomTransfer()
	transfer.FromAccountID = fromAccount.ID
	transfer.ToAccountID = toAccount.ID
	transfer.Amount = amount

	fromEntry := randomEntry()
	fromEntry.AccountID = fromAccount.ID
	fromEntry.Amount = -amount

	toEntry := randomEntry()
	toEntry.AccountID = toAccount.ID
	toEntry.Amount = amount

	fromAccount.Balance -= amount
	toAccount.Balance += amount

	return db.TransferTxResult{
		Transfer:    transfer,
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		FromEntry:   fromEntry,
		ToEntry:     toEntry,
	}
}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfer db.Transfer
	err = json.Unmarshal(data, &gotTransfer)
	require.NoError(t, err)
	require.Equal(t, transfer.ID, gotTransfer.ID)
	require.Equal(t, transfer.FromAccountID, gotTransfer.FromAccountID)
	require.Equal(t, transfer.ToAccountID, gotTransfer.ToAccountID)
	require.Equal(t, transfer.Amount, gotTransfer.Amount)
	require.WithinDuration(t, transfer.CreatedAt, gotTransfer.CreatedAt, time.Second)
}

func requireBodyMatchTransfers(t *testing.T, body *bytes.Buffer, transfers []db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfers []db.Transfer
	err = json.Unmarshal(data, &gotTransfers)
	require.NoError(t, err)
	require.Len(t, gotTransfers, len(transfers))
	for i := range transfers {
		require.Equal(t, transfers[i].ID, gotTransfers[i].ID)
		require.Equal(t, transfers[i].FromAccountID, gotTransfers[i].FromAccountID)
		require.Equal(t, transfers[i].ToAccountID, gotTransfers[i].ToAccountID)
		require.Equal(t, transfers[i].Amount, gotTransfers[i].Amount)
		require.WithinDuration(t, transfers[i].CreatedAt, gotTransfers[i].CreatedAt, time.Second)
	}
}

func requireBodyMatchTransferTxResult(t *testing.T, body *bytes.Buffer, result db.TransferTxResult) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var got db.TransferTxResult
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, result.Transfer.ID, got.Transfer.ID)
	require.Equal(t, result.Transfer.FromAccountID, got.Transfer.FromAccountID)
	require.Equal(t, result.Transfer.ToAccountID, got.Transfer.ToAccountID)
	require.Equal(t, result.Transfer.Amount, got.Transfer.Amount)
	require.WithinDuration(t, result.Transfer.CreatedAt, got.Transfer.CreatedAt, time.Second)

	require.Equal(t, result.FromAccount.ID, got.FromAccount.ID)
	require.Equal(t, result.FromAccount.Owner, got.FromAccount.Owner)
	require.Equal(t, result.FromAccount.Balance, got.FromAccount.Balance)
	require.Equal(t, result.FromAccount.Currency, got.FromAccount.Currency)

	require.Equal(t, result.ToAccount.ID, got.ToAccount.ID)
	require.Equal(t, result.ToAccount.Owner, got.ToAccount.Owner)
	require.Equal(t, result.ToAccount.Balance, got.ToAccount.Balance)
	require.Equal(t, result.ToAccount.Currency, got.ToAccount.Currency)

	require.Equal(t, result.FromEntry.ID, got.FromEntry.ID)
	require.Equal(t, result.FromEntry.AccountID, got.FromEntry.AccountID)
	require.Equal(t, result.FromEntry.Amount, got.FromEntry.Amount)
	require.WithinDuration(t, result.FromEntry.CreatedAt, got.FromEntry.CreatedAt, time.Second)

	require.Equal(t, result.ToEntry.ID, got.ToEntry.ID)
	require.Equal(t, result.ToEntry.AccountID, got.ToEntry.AccountID)
	require.Equal(t, result.ToEntry.Amount, got.ToEntry.Amount)
	require.WithinDuration(t, result.ToEntry.CreatedAt, got.ToEntry.CreatedAt, time.Second)
}
