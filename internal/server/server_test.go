package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chammond14/muzz/internal/db"
)

func Test_loginHandlerReturnsErrorWhenInvalidCredentials(t *testing.T) {
	body := &LoginRequest{Username: "john", Password: "dolphins"}
	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", &bytes)
	res := httptest.NewRecorder()

	TestServer.loginHandler(res, req)

	resBody := &ServerError{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 but got %d", res.Code)
	}

	if resBody.Error != "could not log in" {
		t.Errorf("Expected could not login error but got %s", resBody.Error)
	}
}

func Test_loginHandlerReturnsSuccessfulLogin(t *testing.T) {
	body := &LoginRequest{Username: "john@muzz.com", Password: "papayas"}
	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", &bytes)
	res := httptest.NewRecorder()

	TestServer.loginHandler(res, req)

	resBody := &LoginResponse{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 but got %d", res.Code)
	}

	if resBody.Token == "" {
		t.Error("Expected session token back but got empty string")
	}
}

func Test_discoverHandlerReturnsValidationErrorForUnknownField(t *testing.T) {
	body := struct {
		MinAge string
		Lat    float32
		Long   float32
	}{"test", -0.111, 0.123}

	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(&body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/discover", &bytes)
	res := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), contextKeyUserId, int32(3))
	req = req.WithContext(ctx)

	TestServer.discoverHandler(res, req)

	resBody := &ServerError{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 but got %d", res.Code)
	}

	if resBody.Error != ErrInvalidRequest.Error() {
		t.Errorf("Expected validation error but got %s", resBody.Error)
	}
}

func Test_discoverHandlerReturnsValidationErrorForOutOfBoundsValue(t *testing.T) {
	body := struct {
		MaxAge int
		Lat    float32
		Long   float32
	}{2000, -0.111, 0.123}

	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(&body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/discover", &bytes)
	res := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), contextKeyUserId, int32(3))
	req = req.WithContext(ctx)

	TestServer.discoverHandler(res, req)

	resBody := &ServerError{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 but got %d", res.Code)
	}

	if resBody.Error != ErrValidationError.Error() {
		t.Errorf("Expected validation error but got %s", resBody.Error)
	}
}

func Test_discoverHandlerReturnsNoProfilesWhenNoneMatch(t *testing.T) {
	body := &DiscoverRequest{MinAge: 100, MaxAge: 100, Lat: -0.123, Long: 123}
	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/discover", &bytes)
	res := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), contextKeyUserId, int32(3))
	req = req.WithContext(ctx)

	TestServer.discoverHandler(res, req)

	resBody := &DiscoverResponse{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 but got %d", res.Code)
	}

	if len(resBody.Results) != 0 {
		t.Errorf("Expected no profiles but length of results was %d", len(resBody.Results))
	}
}

func Test_discoverHandlerReturnsProfiles(t *testing.T) {
	body := &DiscoverRequest{MinAge: 18, MaxAge: 100, Lat: -0.123, Long: 123}
	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/discover", &bytes)
	res := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), contextKeyUserId, int32(3))
	req = req.WithContext(ctx)

	TestServer.discoverHandler(res, req)

	resBody := &DiscoverResponse{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 but got %d", res.Code)
	}

	if len(resBody.Results) == 0 {
		t.Errorf("Expected profiles but length of results was %d", len(resBody.Results))
	}
}

func Test_swipeHandlerReturnsNoMatchIdWhenNoMatch(t *testing.T) {
	body := &SwipeRequest{UserId: 1, Liked: true}
	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/swipe", &bytes)
	res := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), contextKeyUserId, int32(3))
	req = req.WithContext(ctx)

	TestServer.swipeHandler(res, req)

	resBody := &SwipeResponse{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 but got %d", res.Code)
	}

	if resBody.Matched != false {
		t.Errorf("Expected no match but got %v", resBody.Matched)
	}

	if resBody.MatchId != 0 {
		t.Errorf("Expected no matchId but got %d", resBody.MatchId)
	}
}

func Test_swipeHandlerReturnsMatchIdWhenMatch(t *testing.T) {
	body := &SwipeRequest{UserId: 3, Liked: true}
	var bytes bytes.Buffer
	err := json.NewEncoder(&bytes).Encode(body)
	if err != nil {
		t.Error("Unexpected error encoding json", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/swipe", &bytes)
	res := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), contextKeyUserId, int32(1))
	req = req.WithContext(ctx)

	TestServer.swipeHandler(res, req)

	resBody := &SwipeResponse{}
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 but got %d", res.Code)
	}

	if resBody.Matched != true {
		t.Errorf("Expected a match but got %v", resBody.Matched)
	}

	if resBody.MatchId == 0 {
		t.Errorf("Expected matchId but got %d", resBody.MatchId)
	}
}

func Test_createUserHandlerReturnsProfile(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/create", nil)
	res := httptest.NewRecorder()

	TestServer.createUserHandler(res, req)

	resBody := &db.Profile{}
	err := json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		t.Error("Unexpected error decoding json", err)
	}

	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 but got %d", res.Code)
	}

	if resBody.Id < 1 {
		t.Error("Expected generated user ID but got zero value")
	}
}
