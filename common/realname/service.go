package realname

import "context"

// Verifier verifies that a legal name matches an identity-card number.
type Verifier interface {
	Verify(ctx context.Context, name, idCardNo string) (*VerifyResult, error)
}

type VerifyResult struct {
	Result      string
	Description string
}
