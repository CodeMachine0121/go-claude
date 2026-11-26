package auth

import "time"

// StatusVo is the sign-in state as normalized by the infrastructure.
type StatusVo struct {
	LoggedIn     bool
	AccountKind  AccountKind
	Plan         string
	ExpiresAt    time.Time
	Email        string
	Organization string
}

// ToDomain converts the observation into the domain model.
func (vo StatusVo) ToDomain() SignInStatusDomain {
	return SignInStatusDomain{status: vo}
}

// SignInVo is a validated sign-in handed to the infrastructure.
type SignInVo struct {
	Method SignInMethod
	Email  string
}
