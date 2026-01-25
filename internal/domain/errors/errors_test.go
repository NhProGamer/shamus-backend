package apperrors_test

import (
	"errors"
	apperrors "shamus-backend/internal/domain/errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *apperrors.AppError
		want string
	}{
		{
			name: "error without wrapped error",
			err:  apperrors.New("TEST_CODE", "test message"),
			want: "TEST_CODE: test message",
		},
		{
			name: "error with wrapped error",
			err:  apperrors.Wrap("TEST_CODE", "test message", errors.New("wrapped error")),
			want: "TEST_CODE: test message (wrapped error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("original error")
	appErr := apperrors.Wrap("TEST_CODE", "test message", wrappedErr)

	unwrapped := appErr.Unwrap()
	if unwrapped != wrappedErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestAppError_Is(t *testing.T) {
	// Test that errors.Is works with AppError
	err := apperrors.ErrGameNotFound

	if !errors.Is(err, apperrors.ErrGameNotFound) {
		t.Error("errors.Is should return true for same error")
	}

	if errors.Is(err, apperrors.ErrPlayerNotFound) {
		t.Error("errors.Is should return false for different error")
	}
}

func TestPredefinedErrors(t *testing.T) {
	// Test that all predefined errors have codes and messages
	tests := []struct {
		name string
		err  *apperrors.AppError
	}{
		{"game not found", apperrors.ErrGameNotFound},
		{"player not found", apperrors.ErrPlayerNotFound},
		{"not your turn", apperrors.ErrNotYourTurn},
		{"wrong phase", apperrors.ErrWrongPhase},
		{"invalid target", apperrors.ErrInvalidTarget},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code == "" {
				t.Errorf("%s: Code should not be empty", tt.name)
			}
			if tt.err.Message == "" {
				t.Errorf("%s: Message should not be empty", tt.name)
			}
		})
	}
}
