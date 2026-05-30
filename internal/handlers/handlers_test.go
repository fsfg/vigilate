package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	tests := []struct {
		name, url, method string
		postedData        url.Values
		handlerFunc       func(http.ResponseWriter, *http.Request)
		expectedStatus    int
	}{
		{
			"Show login screen",
			"/",
			http.MethodGet,
			nil,
			Repo.LoginScreen,
			http.StatusOK,
		},
		{
			"Log user in",
			"/",
			http.MethodPost,
			url.Values{
				"email":    {"me@here.com"},
				"password": {"password"},
			},
			Repo.Login,
			http.StatusSeeOther,
		},
	}

	for _, e := range tests {
		t.Run(fmt.Sprintf("%s:%s", e.method, e.name), func(t *testing.T) {
			t.Parallel()

			var body io.Reader

			if e.postedData != nil {
				body = strings.NewReader(e.postedData.Encode())
			}

			// create a request with body
			req, _ := http.NewRequest(
				e.method,
				e.url,
				body,
			)

			// add the session info to the context
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// create a recorder
			rr := httptest.NewRecorder()

			// cast handler we want to test to an http.HandlerFunc
			handler := http.HandlerFunc(e.handlerFunc)

			// call the handler with our response recorder (which satisfies the response writer interface),
			// and our request (which has our test session). This executes the method we want to test.
			handler.ServeHTTP(rr, req)

			// check returned status code against expected status code
			if rr.Code != e.expectedStatus {
				t.Errorf("%s, expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
			}
		})

	}
}

func TestDBRepo_PusherAuth(t *testing.T) {
	t.Run("Authenticate user to pusher server", func(t *testing.T) {
		postedData := url.Values{
			"socket_id":    {"62406431.2065621642"},
			"channel_name": {"private-channel-2"},
		}

		req, _ := http.NewRequest(
			http.MethodPost,
			"/pusher/auth",
			strings.NewReader(postedData.Encode()),
		)

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		http.HandlerFunc(Repo.PusherAuth).
			ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected response 200, but got %d", rr.Code)
		}

		type pusherResp struct {
			Auth        string `json:"auth"`
			ChannelData string `json:"channel_data"`
		}

		var p pusherResp

		if err := json.NewDecoder(rr.Body).Decode(&p); err != nil {
			t.Fatal(err)
		}

		if len(p.Auth) == 0 {
			t.Error("empty json response")
		}
	})
}
