package commitplan

import (
	"cloud.google.com/go/spanner"
)

// Plan represents a commit plan with mutations to be applied atomically
type Plan struct {
	mutations []*spanner.Mutation
}

// NewPlan creates a new commit plan
func NewPlan() *Plan {
	return &Plan{
		mutations: make([]*spanner.Mutation, 0),
	}
}

// Add adds a mutation to the plan
func (p *Plan) Add(mut *spanner.Mutation) {
	if mut != nil {
		p.mutations = append(p.mutations, mut)
	}
}

// Mutations returns all mutations in the plan
func (p *Plan) Mutations() []*spanner.Mutation {
	return p.mutations
}

// Size returns the number of mutations in the plan
func (p *Plan) Size() int {
	return len(p.mutations)
}
