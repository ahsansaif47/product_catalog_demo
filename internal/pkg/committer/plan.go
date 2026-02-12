package committer

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/pkg/commitplan"
)

// Committer applies commit plans atomically
type Committer struct {
	client *spanner.Client
}

// NewCommitter creates a new committer for Spanner
func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{
		client: client,
	}
}

// Apply applies a commit plan atomically
func (c *Committer) Apply(ctx context.Context, plan *commitplan.Plan) error {
	mutations := plan.Mutations()

	if len(mutations) == 0 {
		return nil
	}

	// Apply all mutations atomically in a single transaction
	_, err := c.client.Apply(ctx, mutations)
	if err != nil {
		return fmt.Errorf("failed to apply commit plan: %w", err)
	}

	return nil
}
