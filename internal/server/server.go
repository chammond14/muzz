package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"

	"github.com/0x6flab/namegenerator"
	"github.com/chammond14/muzz/internal/db"
	"github.com/go-playground/validator/v10"
)

type ServerError struct {
	Error string `json:"error"`
}

type ServerHandler func(http.ResponseWriter, *http.Request)

type Server struct {
	Validate  *validator.Validate
	Generator namegenerator.NameGenerator
	Store     db.ProfileStore
}

var genders = []string{"male", "female", "other"}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type DiscoverRequest struct {
	MinAge  int      `json:"minAge" validate:"omitempty,min=18,max=150"`
	MaxAge  int      `json:"maxAge" validate:"omitempty,min=18,max=150"`
	Genders []string `json:"genders" validate:"dive,oneof=male female other"`
	Lat     float64  `json:"lat" validate:"required"`
	Long    float64  `json:"long" validate:"required"`
}

type DiscoverResponse struct {
	Results []*db.DiscoverProfile `json:"results"`
}

type SwipeRequest struct {
	UserId int32 `json:"user" validate:"required"`
	Liked  bool  `json:"liked"`
}

type SwipeResponse struct {
	Matched bool `json:"matched"`
	MatchId int  `json:"matchId,omitempty"`
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request Received", "Handler", "loginHandler")

	loginRequest, err := createRequestBodyFromRequest(r, &LoginRequest{})
	if err != nil {
		slog.Info("Could not decode request body", "Handler", "loginHandler", "error", err)
		writeErrorResponse(w, err)
		return
	}

	err = s.validateRequest("loginHandler", loginRequest)
	if err != nil {
		slog.Info("error validating request params", "handler", "loginHandler", "error", err)
		writeErrorResponse(w, err)
		return
	}

	sessionToken, err := s.Store.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		slog.Info("Could not login user", "user", loginRequest.Username, "error", err)
		writeErrorResponse(w, err)
		return
	}

	slog.Info("Request Complete", "Handler", "loginHandler")
	writeJsonResponse(w, http.StatusOK, LoginResponse{Token: sessionToken})
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request Received", "Handler", "createUserHandler")

	name := s.Generator.Generate()
	email := fmt.Sprintf("%s@muzz.com", name)
	password := s.Generator.Generate()
	age := rand.IntN(100)
	gender := genders[rand.IntN(len(genders))]
	location := db.Location{Lat: -0.08768348444653988, Long: 51.508050972200834}

	profile, err := s.Store.CreateProfile(r.Context(), age, name, gender, email, password, location)
	if err != nil {
		slog.Info("Error creating profile", "error", err)
		writeErrorResponse(w, err)
		return
	}

	slog.Info("Request Complete", "Handler", "createUserHandler")
	writeJsonResponse(w, http.StatusOK, profile)
}

func (s *Server) discoverHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request Received", "Handler", "discoverHandler")

	discoverRequest, err := createRequestBodyFromRequest(r, &DiscoverRequest{})
	if err != nil {
		slog.Info("Could not decode request body", "Handler", "discoverHandler", "error", err)
		writeErrorResponse(w, ErrInvalidRequest)
		return
	}

	err = s.validateRequest("discoverHandler", discoverRequest)
	if err != nil {
		slog.Info("error validating request params", "handler", "discoverHandler", "error", err)
		writeErrorResponse(w, ErrValidationError)
		return
	}

	userId := r.Context().Value(contextKeyUserId).(int32)
	dbFilters := &db.DiscoverFilters{
		Genders: discoverRequest.Genders,
		MaxAge:  discoverRequest.MaxAge,
		MinAge:  discoverRequest.MinAge,
	}

	discoverResults, err := s.Store.GetDiscoverProfiles(r.Context(), userId, *dbFilters)
	if err != nil {
		slog.Info("Could not load discover profiles", "Handler", "discoverHandler", "error", err)
		writeErrorResponse(w, ErrUnexpectedError)
		return
	}

	userLocation := db.Location{Lat: discoverRequest.Lat, Long: discoverRequest.Long}
	sortProfilesByLocation(discoverResults, userLocation)

	slog.Info("Request Complete", "Handler", "discoverHandler")
	writeJsonResponse(w, http.StatusOK, DiscoverResponse{Results: discoverResults})
}

func (s *Server) swipeHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request Received", "Handler", "swipeHandler")

	swipeRequest, err := createRequestBodyFromRequest(r, &SwipeRequest{})
	if err != nil {
		slog.Info("Could not decode request body", "Handler", "swipeHandler", "error", err)
		writeErrorResponse(w, ErrInvalidRequest)
		return
	}

	err = s.validateRequest("swipeHandler", swipeRequest)
	if err != nil {
		slog.Info("error validating request params", "handler", "swipeHandler", "error", err)
		writeErrorResponse(w, ErrValidationError)
		return
	}

	userId := r.Context().Value(contextKeyUserId).(int32)
	swipeResult, matchId, err := s.Store.Swipe(r.Context(), userId, swipeRequest.UserId, swipeRequest.Liked)
	if err != nil {
		slog.Info("Could not swipe on user", "Handler", "swipeHandler", "error", err)
		writeErrorResponse(w, err)
		return
	}

	response := &SwipeResponse{Matched: swipeResult}
	if swipeResult {
		response.MatchId = matchId
	}

	slog.Info("Request Complete", "Handler", "swipeHandler")
	writeJsonResponse(w, http.StatusOK, response)
}

func (s *Server) Start(addr string) {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /user/create", s.createUserHandler)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.HandleFunc("POST /discover", s.authenticate(s.discoverHandler))
	mux.HandleFunc("POST /swipe", s.authenticate(s.swipeHandler))

	slog.Info("Running on port", "ADDRESS", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Info("Server stopped listening", "error", err)
	}
}

func (s *Server) validateRequest(caller string, target interface{}) error {
	err := s.Validate.Struct(target)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			slog.Error("Validation Error", "Method", caller, "Field", err.Field(), "Tag", err.Tag(), "Value", err.Value())
		}

		return ErrValidationError
	}

	return nil
}

func createRequestBodyFromRequest[T interface{}](request *http.Request, target T) (T, error) {
	err := json.NewDecoder(request.Body).Decode(target)
	if err != nil {
		return target, ErrInvalidRequest
	}

	return target, nil
}

func writeErrorResponse(w http.ResponseWriter, err error) {
	var status int

	switch err {
	case ErrValidationError:
		status = http.StatusBadRequest
	case ErrMustBeLoggedIn:
		status = http.StatusUnauthorized
	case ErrInvalidRequest:
		status = http.StatusBadRequest
	default:
		status = http.StatusInternalServerError
	}

	writeJsonResponse(w, status, ServerError{err.Error()})
}

func writeJsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
