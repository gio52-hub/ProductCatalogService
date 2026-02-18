package committer

import (
	"context"

	"cloud.google.com/go/spanner"
)

// Plan collects Spanner mutations for atomic application.
// This implements a simple version of the Unit of Work pattern.
type Plan struct {
	mutations []*spanner.Mutation
}

// NewPlan creates a new empty Plan.
func NewPlan() *Plan {
	return &Plan{
		mutations: make([]*spanner.Mutation, 0),
	}
}

// Add adds a mutation to the plan.
// Nil mutations are ignored.
func (p *Plan) Add(mut *spanner.Mutation) {
	if mut != nil {
		p.mutations = append(p.mutations, mut)
	}
}

// AddAll adds multiple mutations to the plan.
func (p *Plan) AddAll(muts ...*spanner.Mutation) {
	for _, mut := range muts {
		p.Add(mut)
	}
}

// Mutations returns all collected mutations.
func (p *Plan) Mutations() []*spanner.Mutation {
	return p.mutations
}

// IsEmpty returns true if the plan has no mutations.
func (p *Plan) IsEmpty() bool {
	return len(p.mutations) == 0
}

// Count returns the number of mutations in the plan.
func (p *Plan) Count() int {
	return len(p.mutations)
}

// Clear removes all mutations from the plan.
func (p *Plan) Clear() {
	p.mutations = make([]*spanner.Mutation, 0)
}

// Committer applies plans to Spanner.
type Committer struct {
	client *spanner.Client
}

// NewCommitter creates a new Committer with the given Spanner client.
func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{client: client}
}

// Apply applies all mutations in the plan atomically within a read-write transaction.
func (c *Committer) Apply(ctx context.Context, plan *Plan) error {
	if plan == nil || plan.IsEmpty() {
		return nil
	}

	_, err := c.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return txn.BufferWrite(plan.Mutations())
	})

	return err
}

// ApplyMutations applies the given mutations atomically.
func (c *Committer) ApplyMutations(ctx context.Context, mutations []*spanner.Mutation) error {
	if len(mutations) == 0 {
		return nil
	}

	_, err := c.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return txn.BufferWrite(mutations)
	})

	return err
}

// Client returns the underlying Spanner client.
func (c *Committer) Client() *spanner.Client {
	return c.client
}
