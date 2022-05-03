package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/wiliamhw/simplebank/db/mock"
	db "github.com/wiliamhw/simplebank/db/sqlc"
	"github.com/wiliamhw/simplebank/token"
	"github.com/wiliamhw/simplebank/util"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

// Will be run before or after any other test functions.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

type baseTestCase struct {
	name          string
	buildStubs    func(store *mockdb.MockStore) // Assert DB query calls.
	setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
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
		server := newTestServer(t, store)
		recorder := httptest.NewRecorder()

		request, err := getRequest()
		require.NoError(t, err)

		if btc.setupAuth != nil {
			btc.setupAuth(t, request, server.tokenMaker)
		}
		server.router.ServeHTTP(recorder, request)
		btc.checkResponse(t, recorder)
	})
}
