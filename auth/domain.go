package auth

import "math/bits"

// SignInStatusDomain interprets the sign-in state reported by Claude Code.
type SignInStatusDomain struct {
	status StatusVo
}

// ToDto converts the state into its public shape. A state that is not signed in
// always reads AccountKindNotSignedIn with no account details.
func (d SignInStatusDomain) ToDto() StatusDto {
	if !d.status.LoggedIn {
		return StatusDto{SignedIn: false, AccountKind: AccountKindNotSignedIn}
	}
	return StatusDto{
		SignedIn:     true,
		AccountKind:  d.status.AccountKind,
		Plan:         d.status.Plan,
		ExpiresAt:    d.status.ExpiresAt,
		Email:        d.status.Email,
		Organization: d.status.Organization,
	}
}

// SignInDomain is a sign-in request whose method has been validated.
type SignInDomain struct {
	method SignInMethod
	email  string
}

// NewSignInDomain applies the default method and refuses combined or undefined methods.
func NewSignInDomain(request LoginRequest) (SignInDomain, error) {
	if request.Method&^(ClaudeSubscription|DeveloperAccount|SingleSignOn) != 0 {
		return SignInDomain{}, ErrUnknownSignInMethod
	}
	if bits.OnesCount8(uint8(request.Method)) > 1 {
		return SignInDomain{}, ErrMultipleSignInMethods
	}
	method := request.Method
	if method == 0 {
		method = ClaudeSubscription
	}
	return SignInDomain{method: method, email: request.Email}, nil
}

// ToVo converts the validated sign-in into the shape handed to the infrastructure.
func (d SignInDomain) ToVo() SignInVo {
	return SignInVo{Method: d.method, Email: d.email}
}
