package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var december24 = time.Date(2026, time.December, 24, 21, 37, 16, 360_000_000, time.UTC)

func TestUseCase_Status(t *testing.T) {
	tests := []struct {
		name          string
		observed      StatusVo
		proxyError    error
		expectedDto   StatusDto
		expectedError error
		expectedText  []string
	}{
		{
			name: "signed in with a claude subscription on the team plan",
			observed: StatusVo{
				LoggedIn: true, AccountKind: AccountKindClaudeSubscription, Plan: "team", ExpiresAt: december24,
				Email: "person@example.com", Organization: "Example",
			},
			expectedDto: StatusDto{
				SignedIn: true, AccountKind: AccountKindClaudeSubscription, Plan: "team", ExpiresAt: december24,
				Email: "person@example.com", Organization: "Example",
			},
		},
		{
			name:        "signed in with a developer account",
			observed:    StatusVo{LoggedIn: true, AccountKind: AccountKindDeveloperAccount},
			expectedDto: StatusDto{SignedIn: true, AccountKind: AccountKindDeveloperAccount},
		},
		{
			name:        "not signed in is an ordinary answer",
			observed:    StatusVo{LoggedIn: false, AccountKind: AccountKind("none"), Plan: "stale"},
			expectedDto: StatusDto{SignedIn: false, AccountKind: AccountKindNotSignedIn},
		},
		{
			name:        "an unfamiliar account kind is kept as reported",
			observed:    StatusVo{LoggedIn: true, AccountKind: AccountKind("workload_identity")},
			expectedDto: StatusDto{SignedIn: true, AccountKind: AccountKind("workload_identity")},
		},
		{
			name:        "a subscription pass is reported as such",
			observed:    StatusVo{LoggedIn: true, AccountKind: AccountKindSubscriptionPass},
			expectedDto: StatusDto{SignedIn: true, AccountKind: AccountKindSubscriptionPass},
		},
		{
			name:        "a developer account key is reported as such",
			observed:    StatusVo{LoggedIn: true, AccountKind: AccountKindDeveloperAccountKey},
			expectedDto: StatusDto{SignedIn: true, AccountKind: AccountKindDeveloperAccountKey},
		},
		{
			name:          "an unreadable answer fails and includes what claude code said",
			proxyError:    &UnreadableStatusError{Output: "garbage"},
			expectedError: ErrUnreadableStatus,
			expectedText:  []string{"could not read sign-in state", "garbage"},
		},
		{
			name:          "giving up is reported as given up",
			proxyError:    fmt.Errorf("claude code was given up: %w", context.Canceled),
			expectedError: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := new(mockSignInProxy)
			proxy.On("ReadStatus", mock.Anything).Return(tt.observed, tt.proxyError)

			dto, err := NewUseCase(proxy).Status(context.Background())

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				for _, text := range tt.expectedText {
					assert.Contains(t, err.Error(), text)
				}
				assert.Equal(t, StatusDto{}, dto)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDto, dto)
			}
			proxy.AssertExpectations(t)
		})
	}
}

func TestUseCase_Login(t *testing.T) {
	signedIn := StatusVo{LoggedIn: true, AccountKind: AccountKindClaudeSubscription, Plan: "max"}
	tests := []struct {
		name           string
		request        LoginRequest
		expectedSignIn *SignInVo
		signInError    error
		readStatus     bool
		observed       StatusVo
		expectedDto    StatusDto
		expectedError  error
		expectedText   string
	}{
		{
			name:           "no method chosen signs in with a claude subscription",
			request:        LoginRequest{},
			expectedSignIn: &SignInVo{Method: ClaudeSubscription},
			readStatus:     true,
			observed:       signedIn,
			expectedDto:    StatusDto{SignedIn: true, AccountKind: AccountKindClaudeSubscription, Plan: "max"},
		},
		{
			name:           "developer account with a pre-filled email",
			request:        LoginRequest{Method: DeveloperAccount, Email: "dev@example.com"},
			expectedSignIn: &SignInVo{Method: DeveloperAccount, Email: "dev@example.com"},
			readStatus:     true,
			observed:       StatusVo{LoggedIn: true, AccountKind: AccountKindDeveloperAccount},
			expectedDto:    StatusDto{SignedIn: true, AccountKind: AccountKindDeveloperAccount},
		},
		{
			name:           "company single sign-on",
			request:        LoginRequest{Method: SingleSignOn},
			expectedSignIn: &SignInVo{Method: SingleSignOn},
			readStatus:     true,
			observed:       signedIn,
			expectedDto:    StatusDto{SignedIn: true, AccountKind: AccountKindClaudeSubscription, Plan: "max"},
		},
		{
			name:           "the person closes the sign-in without finishing",
			request:        LoginRequest{},
			expectedSignIn: &SignInVo{Method: ClaudeSubscription},
			signInError:    ErrSignInNotCompleted,
			expectedError:  ErrSignInNotCompleted,
			expectedText:   "sign-in was not completed",
		},
		{
			name:           "the sign-in ends but the person is still signed out",
			request:        LoginRequest{},
			expectedSignIn: &SignInVo{Method: ClaudeSubscription},
			readStatus:     true,
			observed:       StatusVo{LoggedIn: false},
			expectedError:  ErrSignInNotCompleted,
			expectedText:   "sign-in was not completed",
		},
		{
			name:          "two methods chosen at once are refused before anything starts",
			request:       LoginRequest{Method: DeveloperAccount | SingleSignOn},
			expectedError: ErrMultipleSignInMethods,
			expectedText:  "choose one sign-in method",
		},
		{
			name:          "an undefined method is refused before anything starts",
			request:       LoginRequest{Method: SignInMethod(8)},
			expectedError: ErrUnknownSignInMethod,
			expectedText:  "unknown sign-in method",
		},
		{
			name:           "giving up on the sign-in stops it",
			request:        LoginRequest{},
			expectedSignIn: &SignInVo{Method: ClaudeSubscription},
			signInError:    fmt.Errorf("claude code was given up: %w", context.Canceled),
			expectedError:  context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := new(mockSignInProxy)
			if tt.expectedSignIn != nil {
				proxy.On("SignIn", mock.Anything, *tt.expectedSignIn).Return(tt.signInError)
			}
			if tt.readStatus {
				proxy.On("ReadStatus", mock.Anything).Return(tt.observed, nil)
			}

			dto, err := NewUseCase(proxy).Login(context.Background(), tt.request)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				assert.Contains(t, err.Error(), tt.expectedText)
				assert.Equal(t, StatusDto{}, dto)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDto, dto)
			}
			proxy.AssertExpectations(t)
			if tt.expectedSignIn == nil {
				proxy.AssertNotCalled(t, "SignIn", mock.Anything, mock.Anything)
			}
			if !tt.readStatus {
				proxy.AssertNotCalled(t, "ReadStatus", mock.Anything)
			}
		})
	}
}

func TestUseCase_Login_UnreadableStateAfterSignIn(t *testing.T) {
	proxy := new(mockSignInProxy)
	proxy.On("SignIn", mock.Anything, SignInVo{Method: ClaudeSubscription}).Return(nil)
	proxy.On("ReadStatus", mock.Anything).Return(StatusVo{}, &UnreadableStatusError{Output: "garbage"})

	_, err := NewUseCase(proxy).Login(context.Background(), LoginRequest{})

	require.ErrorIs(t, err, ErrUnreadableStatus)
}

func TestUseCase_Logout(t *testing.T) {
	tests := []struct {
		name          string
		signOutError  error
		expectedError error
	}{
		{name: "signing out succeeds, including when already signed out"},
		{name: "a failed sign-out is reported", signOutError: ErrSignOutFailed, expectedError: ErrSignOutFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := new(mockSignInProxy)
			proxy.On("SignOut", mock.Anything).Return(tt.signOutError)

			err := NewUseCase(proxy).Logout(context.Background())

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
			proxy.AssertExpectations(t)
		})
	}
}
