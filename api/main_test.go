package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/wiliamhw/simplebank/db/mock"
)

// Will be run before or after any other test functions.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
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
