package websocket

import (
	"errors"
	"shamus-backend/internal/domain/entities"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/olahol/melody"
)

// Session extraction errors
var (
	ErrMissingGameID   = errors.New("missing gameId in session")
	ErrMissingUserID   = errors.New("missing userId in session")
	ErrMissingUserInfo = errors.New("missing userInfo in session")
	ErrInvalidGameID   = errors.New("invalid gameId type in session")
	ErrInvalidUserID   = errors.New("invalid userId type in session")
	ErrInvalidUserInfo = errors.New("invalid userInfo type in session")
)

// SessionData holds extracted and validated session data
type SessionData struct {
	GameID   entities.GameID
	PlayerID entities.PlayerID
	UserInfo *oidc.UserInfo
	Username string
}

// ExtractGameID safely extracts the game ID from a melody session
func ExtractGameID(s *melody.Session) (entities.GameID, error) {
	val, exists := s.Get("gameId")
	if !exists || val == nil {
		return "", ErrMissingGameID
	}
	str, ok := val.(string)
	if !ok {
		return "", ErrInvalidGameID
	}
	if str == "" {
		return "", ErrMissingGameID
	}
	return entities.GameID(str), nil
}

// ExtractPlayerID safely extracts the player ID from a melody session
func ExtractPlayerID(s *melody.Session) (entities.PlayerID, error) {
	val, exists := s.Get("userId")
	if !exists || val == nil {
		return "", ErrMissingUserID
	}
	str, ok := val.(string)
	if !ok {
		return "", ErrInvalidUserID
	}
	if str == "" {
		return "", ErrMissingUserID
	}
	return entities.PlayerID(str), nil
}

// ExtractUserInfo safely extracts the OIDC user info from a melody session
func ExtractUserInfo(s *melody.Session) (*oidc.UserInfo, error) {
	val, exists := s.Get("userInfo")
	if !exists || val == nil {
		return nil, ErrMissingUserInfo
	}
	userInfo, ok := val.(oidc.UserInfo)
	if !ok {
		// Also try pointer type
		userInfoPtr, okPtr := val.(*oidc.UserInfo)
		if !okPtr {
			return nil, ErrInvalidUserInfo
		}
		return userInfoPtr, nil
	}
	return &userInfo, nil
}

// ExtractUsername extracts the username from user info, with fallback to player ID
func ExtractUsername(userInfo *oidc.UserInfo, playerID entities.PlayerID) string {
	if userInfo == nil {
		return string(playerID)
	}
	var claims struct {
		Username string `json:"preferred_username"`
	}
	if err := userInfo.Claims(&claims); err != nil || claims.Username == "" {
		return string(playerID)
	}
	return claims.Username
}

// ExtractSessionData extracts all session data with safe type assertions
func ExtractSessionData(s *melody.Session) (*SessionData, error) {
	gameID, err := ExtractGameID(s)
	if err != nil {
		return nil, err
	}

	playerID, err := ExtractPlayerID(s)
	if err != nil {
		return nil, err
	}

	userInfo, err := ExtractUserInfo(s)
	if err != nil {
		return nil, err
	}

	username := ExtractUsername(userInfo, playerID)

	return &SessionData{
		GameID:   gameID,
		PlayerID: playerID,
		UserInfo: userInfo,
		Username: username,
	}, nil
}
