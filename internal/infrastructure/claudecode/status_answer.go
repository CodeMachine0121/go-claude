package claudecode

import "encoding/json"

// identityFields are the parts of a sign-in answer that identify the person.
var identityFields = []string{"email", "orgId", "orgName", "accountUuid", "organizationUuid"}

// statusAnswer is the text Claude Code printed for `claude auth status`.
type statusAnswer struct {
	text string
}

// WithoutIdentity returns the answer with the person's identity removed, so it
// can be shown in a failure. Text that is not a structured answer carries no
// identity fields and is returned as-is.
func (a statusAnswer) WithoutIdentity() string {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(a.text), &fields); err != nil {
		return a.text
	}
	for _, name := range identityFields {
		delete(fields, name)
	}
	// Re-encoding raw messages that were just decoded cannot fail.
	scrubbed, _ := json.Marshal(fields)
	return string(scrubbed)
}
