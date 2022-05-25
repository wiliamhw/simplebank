package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/wiliamhw/simplebank/db/mock"
	db "github.com/wiliamhw/simplebank/db/sqlc"
	"github.com/wiliamhw/simplebank/util"
)

const userURI = "/users"

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	defaultBody := gin.H{
		"username":  user.Username,
		"password":  password,
		"full_name": user.FullName,
		"email":     user.Email,
	}

	testCases := []struct {
		base baseTestCase
		body gin.H
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					}

					store.EXPECT().
						CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
						Times(1).
						Return(user, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
					requireBodyMatchUser(t, recorder.Body, user)
				},
			},
			body: defaultBody,
		},
		{
			base: baseTestCase{
				name: "InternalError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.User{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			body: defaultBody,
		},
		{
			base: baseTestCase{
				name: "DuplicateUsername",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.User{}, &pq.Error{Code: "23505"})
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusForbidden, recorder.Code)
				},
			},
			body: defaultBody,
		},
		{
			base: baseTestCase{
				name: "InvalidUsername",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			body: gin.H{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
		},
		{
			base: baseTestCase{
				name: "InvalidEmail",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
		},
		{
			base: baseTestCase{
				name: "TooShortPassword",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			body: gin.H{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			if err != nil {
				return nil, err
			}

			request, err := http.NewRequest(
				http.MethodPost, userURI, bytes.NewReader(data),
			)
			if err != nil {
				return nil, err
			}
			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func TestLoginUserAPI(t *testing.T) {
	user, password := randomUser(t)

	defaultBody := gin.H{
		"username": user.Username,
		"password": password,
	}

	testCases := []struct {
		base baseTestCase
		body gin.H
	}{
		{
			base: baseTestCase{
				name: "OK",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)

					store.EXPECT().
						CreateSession(gomock.Any(), gomock.Any()).
						Times(1)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusOK, recorder.Code)
				},
			},
			body: defaultBody,
		},
		{
			base: baseTestCase{
				name: "UserNotFound",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.User{}, sql.ErrNoRows)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusNotFound, recorder.Code)
				},
			},
			body: gin.H{
				"username": "NotFound",
				"password": password,
			},
		},
		{
			base: baseTestCase{
				name: "IncorrectPassword",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusUnauthorized, recorder.Code)
				},
			},
			body: gin.H{
				"username": user.Username,
				"password": "incorrect",
			},
		},
		{
			base: baseTestCase{
				name: "InternalError",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(1).
						Return(db.User{}, sql.ErrConnDone)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusInternalServerError, recorder.Code)
				},
			},
			body: defaultBody,
		},
		{
			base: baseTestCase{
				name: "InvalidUsername",
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				},
			},
			body: gin.H{
				"username": "invalid-user#1",
				"password": password,
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		getRequest := func() (*http.Request, error) {
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			if err != nil {
				return nil, err
			}

			request, err := http.NewRequest(
				http.MethodPost, userURI+"/login", bytes.NewReader(data),
			)
			if err != nil {
				return nil, err
			}
			return request, err
		}
		tc.base.runTestCase(t, getRequest)
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
