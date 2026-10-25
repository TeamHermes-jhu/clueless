package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	"github.com/TeamHermes-jhu/clueless/backend/session"
)

func (s *Server) joinSessionHandler(w http.ResponseWriter, r *http.Request) {
	req, err := parseJoinRequest(r)
	if err != nil {
		http.Error(w, "malformed request", http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sess := range s.sessions {
		if !sess.Add(req.GetUsername()) {
			continue
		}

		respBody, err := generateJoinResponse(sess.GetID().String(), req.GetUsername())
		if err != nil {
			http.Error(w, "failed to marshal response", http.StatusInternalServerError)

			return
		}

		_, _ = w.Write(respBody)

		return
	}

	if len(s.sessions) >= maxSessionsCount {
		_, _ = w.Write([]byte("Sorry, all sessions full"))

		return
	}

	sess := session.New(s.logger)
	s.sessions[sess.GetID()] = sess
	sess.Add(req.GetUsername())

	respBody, err := generateJoinResponse(sess.GetID().String(), req.GetUsername())
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)

		return
	}

	_, _ = w.Write(respBody)
}

func parseJoinRequest(r *http.Request) (*sessionv1.JoinRequest, error) {
	var req sessionv1.JoinRequest

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(body, &req)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if !req.HasUsername() {
		return nil, errors.New("required: username")
	}

	return &req, nil
}

func generateJoinResponse(sessionID, player string) ([]byte, error) {
	resp := sessionv1.JoinResponse_builder{
		Username:  new(player),
		SessionId: new(sessionID),
	}.Build()

	return protojson.Marshal(resp)
}
