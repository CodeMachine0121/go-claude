package claudecode

import "github.com/CodeMachine0121/go-claude/auth"

// signInCommand translates a validated sign-in into `claude auth login` arguments.
type signInCommand struct {
	signIn auth.SignInVo
}

func (c signInCommand) arguments() []string {
	arguments := []string{"auth", "login"}
	switch c.signIn.Method {
	case auth.ClaudeSubscription:
		arguments = append(arguments, "--claudeai")
	case auth.DeveloperAccount:
		arguments = append(arguments, "--console")
	case auth.SingleSignOn:
		arguments = append(arguments, "--sso")
	}
	if c.signIn.Email != "" {
		arguments = append(arguments, "--email", c.signIn.Email)
	}
	return arguments
}
