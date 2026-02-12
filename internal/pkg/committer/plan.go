package committer

import (
	"context"

	"cloud.google.com/go/spanner"
	commitplan "github.com/Vektor-AI/commitplan"
	commitplanSpanner "github.com/Vektor-AI/commitplan/drivers/spanner"
)

// Committer applies commit plans
type Committer struct {
	driver *commitplanSpanner.Driver
}

// NewCommitter creates a new committer with Spanner driver
func NewCommitter(client *spanner.Client) *Committer {
	driver := commitplanSpanner.NewDriver(client)

	return &Committer{
		driver: driver,
	}
}

// Apply applies a commit plan atomically
func (c *Committer) Apply(ctx context.Context, plan *commitplan.Plan) error {
	// Convert our plan to the actual commitplan type and apply
	return c.driver.Apply(ctx, plan)
}
