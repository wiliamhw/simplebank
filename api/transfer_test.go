package api

// const transferURI = "/transfers"

// func TestTransferAPI(t *testing.T) {
// 	amount := int64(10)

// 	account1 := randomAccount(t)
// 	account2 := randomAccount(t)
// 	account3 := randomAccount(t)

// 	account1.Currency = util.USD
// 	account2.Currency = util.USD
// 	account3.Currency = util.EUR

// 	testCases := []struct {
// 		base baseTestCase
// 		body gin.H
// 	}{
// 		{
// 			base: baseTestCase{
// 				name: "OK",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().
// 						GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
// 						Times(1).
// 						Return(account1, nil)

// 					store.EXPECT().
// 						GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
// 						Times(1).
// 						Return(account2, nil)

// 					arg := db.TransferTxParams{
// 						FromAccountID: account1.ID,
// 						ToAccountID:   account2.ID,
// 						Amount:        amount,
// 					}
// 					store.EXPECT().
// 						TransferTx(gomock.Any(), gomock.Eq(arg)).
// 						Times(1)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusOK, recorder.Code)
// 				},
// 			},
// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "FromAccountNotFound",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusNotFound, recorder.Code)
// 				},
// 			},
// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "ToAccountNotFound",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusNotFound, recorder.Code)
// 				},
// 			},
// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "FromAccountCurrencyMismatch",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusBadRequest, recorder.Code)
// 				},
// 			},

// 			body: gin.H{
// 				"from_account_id": account3.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "ToAccountCurrencyMismatch",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusBadRequest, recorder.Code)
// 				},
// 			},
// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account3.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "InvalidCurrency",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusBadRequest, recorder.Code)
// 				},
// 			},

// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        "XYZ",
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "NegativeAmount",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusBadRequest, recorder.Code)
// 				},
// 			},

// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          -amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "GetAccountError",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, sql.ErrConnDone)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 				},
// 			},

// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 		{
// 			base: baseTestCase{
// 				name: "TransferTxError",
// 				buildStubs: func(store *mockdb.MockStore) {
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
// 					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
// 					store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, sql.ErrTxDone)
// 				},
// 				checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 					require.Equal(t, http.StatusInternalServerError, recorder.Code)
// 				},
// 			},

// 			body: gin.H{
// 				"from_account_id": account1.ID,
// 				"to_account_id":   account2.ID,
// 				"amount":          amount,
// 				"currency":        util.USD,
// 			},
// 		},
// 	}

// 	for i := range testCases {
// 		tc := testCases[i]

// 		getRequest := func() (*http.Request, error) {
// 			// Marshal body data to JSON
// 			data, err := json.Marshal(tc.body)
// 			if err != nil {
// 				return nil, err
// 			}

// 			request, err := http.NewRequest(
// 				http.MethodPost, transferURI, bytes.NewReader(data),
// 			)
// 			if err != nil {
// 				return nil, err
// 			}

// 			return request, err
// 		}
// 		tc.base.runTestCase(t, getRequest)
// 	}
// }
