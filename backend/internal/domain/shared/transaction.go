package shared

import "context"

// TransactionManager abstracts database transaction boundaries.
// Infrastructure layers provide the concrete implementation.
type TransactionManager interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}
