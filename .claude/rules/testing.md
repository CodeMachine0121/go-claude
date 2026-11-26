# Testing

## Principles

- **Test business behavior only**: given input → expected business outcome. Never assert
  which unexported methods ran or how fields are stored.
- Drive the subject **only through exported methods**. Refactoring internals must not
  break tests.
- Do not test the standard library, the Go toolchain or the Claude Code CLI itself.
- **Table-driven tests** with `t.Run()` subtests; one case per input/output pair. Cover
  the happy path, every boundary and every error.

## Where the weight goes

| Subject | How | Mock what |
| :--- | :--- | :--- |
| `UseCase` | **Main focus**: happy path, boundaries, every error | Only the unexported `iXxxProxy` |
| Domain models / VOs | Direct unit tests: inputs in, results out | Nothing |
| Composition root (`claude.New`) | Option validation: conflicts, defaults | Nothing; no external resources |
| `ClaudeCodeProxy` process I/O | No unit tests; covered by integration tests | — |
| Pure code inside infrastructure (wire decoding, secret redaction, argument building) | Unit tests in the same package, fed with **recorded real CLI output** pasted as test data. Recordings are data, not fakes: they carry no logic | Nothing |
| Real Claude Code CLI | Integration tests in `test/` | Nothing |

## Mocking

- **Only mock interfaces, only with `testify/mock`.**
- **No hand-written fakes**: no `fakeProxy`, no in-memory stand-ins, no fake CLI
  executables or scripts with their own logic.
- A mock embeds `mock.Mock` and its methods only call `m.Called(...)`; no custom logic.
- Mocks sit only at the outermost boundary (the proxy). Internal business components are
  injected as real instances — this deliberately exercises the logic beneath.
- Each subtest creates its own mock.

## Tools

- Assertions: `github.com/stretchr/testify/assert` and `require`.
- Mocks: `github.com/stretchr/testify/mock`.
- testify appears in `_test.go` files only; it is never a production dependency.

## File placement

| File | Location | Notes |
| :--- | :--- | :--- |
| Unit tests `*_test.go` | Same feature folder, **same package** | Needed to reach the unexported proxy interface. Still only call exported methods. |
| Mocks | `mock_test.go` in the feature folder | Test-only compilation keeps mocks out of the public API |
| Runnable examples | `example_test.go`, external `<pkg>_test` package | Shown in godoc. Examples that need no external resource carry `// Output:` so they run |
| Integration tests | `test/`, first line `//go:build integration` | `go test -tags integration ./test/...` |

## Integration tests

- **Always hit the real Claude Code CLI.** Never substitute a fake binary or server.
- When the resource is unavailable (not installed, not signed in), `t.Skip` with the reason
  instead of failing.
- Operations that change the user's state (signing out) never run against the user's
  everyday sign-in: use an isolated profile directory created by the test.

## Template

### `auth/mock_test.go`

```go
package auth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockSignInProxy struct {
	mock.Mock
}

func (m *mockSignInProxy) ReadStatus(ctx context.Context) (StatusVo, error) {
	args := m.Called(ctx)
	return args.Get(0).(StatusVo), args.Error(1)
}
```

### `auth/usecase_test.go`

```go
func TestUseCase_Status(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(m *mockSignInProxy)
		expectedDto   StatusDto
		expectedError error
	}{
		{
			name: "signed in",
			mockSetup: func(m *mockSignInProxy) {
				m.On("ReadStatus", mock.Anything).Return(StatusVo{LoggedIn: true}, nil)
			},
			expectedDto: StatusDto{SignedIn: true},
		},
		{
			name: "CLI not installed",
			mockSetup: func(m *mockSignInProxy) {
				m.On("ReadStatus", mock.Anything).Return(StatusVo{}, ErrNotInstalled)
			},
			expectedError: ErrNotInstalled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProxy := new(mockSignInProxy)
			tt.mockSetup(mockProxy)
			useCase := NewUseCase(mockProxy)

			dto, err := useCase.Status(context.Background())

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDto, dto)
			}
			mockProxy.AssertExpectations(t)
		})
	}
}
```
