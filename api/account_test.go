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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/wiliamhw/simplebank/db/mock"
	db "github.com/wiliamhw/simplebank/db/sqlc"
	"github.com/wiliamhw/simplebank/token"
	"github.com/wiliamhw/simplebank/util"
)

const accountURI = "/accounts"

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		base      baseTestCase
		accountId int64
	}{
		{
			base: baseTestCase{
				name: "OK",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
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
				name: "UnauthorizedUser",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(account, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "NoAuthorization",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "NotFound",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
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
			url := fmt.Sprintf("%v/%v", accountURI, tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}

			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func TestListAccountAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	var accounts []db.Account = make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(user.Username)
	}

	defaultRequest := listAccountRequest{
		PageID:   1,
		PageSize: int32(n),
	}

	testCases := []struct {
		base baseTestCase
		req  listAccountRequest
	}{
		{
			base: baseTestCase{
				name: "OK",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  int32(n),
						Offset: 0,
					}
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

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
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "UnauthorizedUser",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq("unauthorized_user")).
						Times(1).
						Return(db.User{}, sql.ErrConnDone)

					store.EXPECT().
						ListAccounts(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "NoAuthorization",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(0)

					store.EXPECT().
						ListAccounts(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "InvalidQueryParams",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(0)

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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  int32(n),
						Offset: 0,
					}
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

					store.EXPECT().
						ListAccounts(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return([]db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			req: defaultRequest,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			request, err := http.NewRequest(http.MethodGet, accountURI, nil)
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

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	createAccountParams := db.CreateAccountParams{
		Owner:    account.Owner,
		Currency: account.Currency,
		Balance:  0,
	}

	defaultRequest := createAccountRequest{
		Currency: account.Currency,
	}

	testCases := []struct {
		base baseTestCase
		req  createAccountRequest
	}{
		{
			base: baseTestCase{
				name: "OK",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Eq(createAccountParams)).
						Times(1).
						Return(account, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchAccount(t, recorder.Body, account)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "UnauthorizedUser",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq("unauthorized_user")).
						Times(1).
						Return(db.User{}, sql.ErrConnDone)

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "NoAuthorization",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "InvalidBody",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			req: createAccountRequest{
				Currency: "RP",
			},
		},
		{
			base: baseTestCase{
				name: "InternalError",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

					store.EXPECT().
						CreateAccount(gomock.Any(), gomock.Eq(createAccountParams)).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			req: defaultRequest,
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
				http.MethodPost, accountURI, bytes.NewReader(data),
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
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	amount := util.RandomInt(-100, 100)
	updatedAccount := account
	updatedAccount.Balance = account.Balance + amount

	defaultRequest := updateAccountRequest{
		uri: updateAccountRequestUri{
			ID: account.ID,
		},
		body: updateAccountRequestForm{
			Amount: amount,
		},
	}

	testCases := []struct {
		base baseTestCase
		req  updateAccountRequest
	}{
		{
			base: baseTestCase{
				name: "OK",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.AddAccountBalanceParams{
						ID:     account.ID,
						Amount: amount,
					}
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(account, nil)

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
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "UnauthorizedUser",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq("unauthorized_user")).
						Times(1).
						Return(db.User{}, sql.ErrConnDone)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "NoAuthorization",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			req: defaultRequest,
		},
		{
			base: baseTestCase{
				name: "InvalidQuery",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(1).
						Return(account, nil)

					store.EXPECT().
						AddAccountBalance(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			req: defaultRequest,
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

			url := fmt.Sprintf("%v/%v", accountURI, tc.req.uri.ID)
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
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		base      baseTestCase
		accountId int64
	}{
		{
			base: baseTestCase{
				name: "OK",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

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
				name: "UnauthorizedUser",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq("unauthorized_user")).
						Times(1).
						Return(db.User{}, sql.ErrConnDone)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "NoAuthorization",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Any()).
						Times(0)

					store.EXPECT().
						DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			accountId: account.ID,
		},
		{
			base: baseTestCase{
				name: "NotFound",
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

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
				setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
					addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)

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
			url := fmt.Sprintf("%v/%v", accountURI, tc.accountId)
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
