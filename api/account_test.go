package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/wiliamhw/simplebank/db/mock"
	db "github.com/wiliamhw/simplebank/db/sqlc"
	"github.com/wiliamhw/simplebank/util"
)

const baseURI = "/accounts"

func randomAccount(t *testing.T) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

type baseTestCase struct {
	name          string
	buildStubs    func(store *mockdb.MockStore) // Assert DB query calls.
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func (btc *baseTestCase) runTestCase(
	t *testing.T,
	getRequest func() (*http.Request, error),
) {
	t.Run(btc.name, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		btc.buildStubs(store)

		// Start test server and send request
		server := NewServer(store)
		recorder := httptest.NewRecorder()

		request, err := getRequest()
		require.NoError(t, err)

		server.router.ServeHTTP(recorder, request)
		btc.checkResponse(t, recorder)
	})
}

func TestListAccountAPI(t *testing.T) {
	n := 5
	var accounts []db.Account = make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(t)
	}

	testCases := []struct {
		base baseTestCase
		req  listAccountRequest
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.ListAccountsParams{
						Limit:  int32(n),
						Offset: 0,
					}

					store.EXPECT().
						ListAccounts(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(accounts, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccounts(t, recorder.Body, accounts)
				},
			},
			req: listAccountRequest{
				PageID:   1,
				PageSize: int32(n),
			},
		},
		{
			base: baseTestCase{
				name: "InvalidQueryParams",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						ListAccounts(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			req: listAccountRequest{
				PageID:   -1,
				PageSize: int32(1000),
			},
		},
		{
			base: baseTestCase{
				name: "InternalError",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.ListAccountsParams{
						Limit:  int32(n),
						Offset: 0,
					}
					store.EXPECT().
						ListAccounts(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return([]db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			req: listAccountRequest{
				PageID:   1,
				PageSize: int32(n),
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			request, err := http.NewRequest(http.MethodGet, baseURI, nil)
			if err != nil {
				return nil, err
			}

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.req.PageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.req.PageSize))
			request.URL.RawQuery = q.Encode()

			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount(t)

	testCases := []struct {
		base      baseTestCase
		accountId int64
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(account, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, account)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "NotFound",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(db.Account{}, sql.ErrNoRows)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusNotFound, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "InternalError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "InvalidID",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			accountId: 0,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			url := fmt.Sprintf("%v/%v", baseURI, tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}

			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func TestCreateAccountAPI(t *testing.T) {
	account := randomAccount(t)

	testCases := []struct {
		base baseTestCase
		req  createAccountRequest
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.CreateAccountParams{
						Owner:    account.Owner,
						Currency: account.Currency,
						Balance:  0,
					}

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(account, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, account)
				},
			},
			req: createAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
		},
		{
			base: baseTestCase{
				name: "InvalidBody",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			req: createAccountRequest{
				Owner:    "./",
				Currency: "RP",
			},
		},
		{
			base: baseTestCase{
				name: "InternalError",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.CreateAccountParams{
						Owner:    account.Owner,
						Currency: account.Currency,
						Balance:  0,
					}

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			req: createAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			// Marshal body data to JSON
			data, err := json.Marshal(tc.req)
			if err != nil {
				return nil, err
			}

			request, err := http.NewRequest(
				http.MethodPost, baseURI, bytes.NewReader(data),
			)
			if err != nil {
				return nil, err
			}

			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func TestUpdateAccountAPI(t *testing.T) {
	account := randomAccount(t)

	amount := util.RandomInt(-100, 100)
	updatedAccount := account
	updatedAccount.Balance = account.Balance + amount

	testCases := []struct {
		base baseTestCase
		req  updateAccountRequest
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.AddAccountBalanceParams{
						ID:     account.ID,
						Amount: amount,
					}

					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(updatedAccount, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, updatedAccount)
				},
			},

			req: updateAccountRequest{
				uri: updateAccountRequestUri{
					ID: account.ID,
				},
				body: updateAccountRequestForm{
					Amount: amount,
				},
			},
		},
		{
			base: baseTestCase{
				name: "InvalidQuery",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},

			req: updateAccountRequest{
				uri: updateAccountRequestUri{
					ID: -1,
				},
				body: updateAccountRequestForm{
					Amount: amount,
				},
			},
		},
		{
			base: baseTestCase{
				name: "InvalidBody",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			req: updateAccountRequest{
				uri: updateAccountRequestUri{
					ID: account.ID,
				},
				body: updateAccountRequestForm{
					Amount: 0,
				},
			},
		},
		{
			base: baseTestCase{
				name: "InternalError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			req: updateAccountRequest{
				uri: updateAccountRequestUri{
					ID: account.ID,
				},
				body: updateAccountRequestForm{
					Amount: amount,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			// Marshal body data to JSON
			data, err := json.Marshal(tc.req.body)
			if err != nil {
				return nil, err
			}

			url := fmt.Sprintf("%v/%v", baseURI, tc.req.uri.ID)
			request, err := http.NewRequest(
				http.MethodPatch, url, bytes.NewReader(data),
			)
			if err != nil {
				return nil, err
			}

			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func TestDeleteAccountAPI(t *testing.T) {
	account := randomAccount(t)

	testCases := []struct {
		base      baseTestCase
		accountId int64
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(account, nil)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "NotFound",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(db.Account{}, sql.ErrNoRows)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusNotFound, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "InternalErrorOnAssertingAccountExistance",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "InternalErrorOnAccountDeletion",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(account, nil)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(1).
						Return(sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "InvalidID",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			accountId: 0,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			url := fmt.Sprintf("%v/%v", baseURI, tc.accountId)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				return nil, err
			}
			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
