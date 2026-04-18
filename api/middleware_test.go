package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KamisAyaka/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/KamisAyaka/simplebank/token"
)

func TestAuthMiddleware(t *testing.T) {
	user, _ := randomUser()

	tokenMaker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	const authPath = "/auth"

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "NoAuthorizationHeader",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationHeaderFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				request.Header.Set(authorizationHeaderKey, "invalid-header-format")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorizationType",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "basic", user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				request.Header.Set(authorizationHeaderKey, "bearer invalid_token")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				var body map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &body)
				require.NoError(t, err)
				require.Equal(t, user.Username, body["username"])
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET(authPath, authMiddleware(tokenMaker), func(ctx *gin.Context) {
				payload, exists := ctx.Get(authorizationPayloadKey)
				require.True(t, exists)

				authPayload, ok := payload.(*token.Payload)
				require.True(t, ok)
				require.Equal(t, user.Username, authPayload.Username)

				ctx.JSON(http.StatusOK, gin.H{"username": authPayload.Username})
			})

			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, tokenMaker)
			router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authorizationHeader := authorizationType + " " + token
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}
