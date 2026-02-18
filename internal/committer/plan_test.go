package committer

import (
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
)

func TestNewPlan(t *testing.T) {
	t.Parallel()

	plan := NewPlan()

	assert.NotNil(t, plan)
	assert.Empty(t, plan.Mutations())
}

func TestPlan_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		mutations         []*spanner.Mutation
		expectedCount     int
		includeNilMut     bool
		expectedNilHandle bool
	}{
		{
			name:          "add single mutation",
			mutations:     []*spanner.Mutation{spanner.Insert("table", []string{"col"}, []interface{}{"val"})},
			expectedCount: 1,
		},
		{
			name: "add multiple mutations",
			mutations: []*spanner.Mutation{
				spanner.Insert("table1", []string{"col"}, []interface{}{"val1"}),
				spanner.Insert("table2", []string{"col"}, []interface{}{"val2"}),
			},
			expectedCount: 2,
		},
		{
			name:              "add nil mutation",
			mutations:         []*spanner.Mutation{nil},
			expectedCount:     0,
			includeNilMut:     true,
			expectedNilHandle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := NewPlan()

			for _, m := range tt.mutations {
				plan.Add(m)
			}

			assert.Equal(t, tt.expectedCount, len(plan.Mutations()))
		})
	}
}

func TestPlan_AddAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mutations     []*spanner.Mutation
		expectedCount int
	}{
		{
			name:          "add empty slice",
			mutations:     []*spanner.Mutation{},
			expectedCount: 0,
		},
		{
			name: "add slice with mutations",
			mutations: []*spanner.Mutation{
				spanner.Insert("table1", []string{"col"}, []interface{}{"val1"}),
				spanner.Insert("table2", []string{"col"}, []interface{}{"val2"}),
				spanner.Insert("table3", []string{"col"}, []interface{}{"val3"}),
			},
			expectedCount: 3,
		},
		{
			name: "add slice with nil mutations",
			mutations: []*spanner.Mutation{
				spanner.Insert("table1", []string{"col"}, []interface{}{"val1"}),
				nil,
				spanner.Insert("table2", []string{"col"}, []interface{}{"val2"}),
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := NewPlan()
			plan.AddAll(tt.mutations...)

			assert.Equal(t, tt.expectedCount, len(plan.Mutations()))
		})
	}
}

func TestPlan_Mutations(t *testing.T) {
	t.Parallel()

	plan := NewPlan()

	m1 := spanner.Insert("table1", []string{"col"}, []interface{}{"val1"})
	m2 := spanner.Insert("table2", []string{"col"}, []interface{}{"val2"})

	plan.Add(m1)
	plan.Add(m2)

	mutations := plan.Mutations()
	assert.Len(t, mutations, 2)
}

func TestPlan_IsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		addMutations bool
		expected     bool
	}{
		{
			name:         "empty plan",
			addMutations: false,
			expected:     true,
		},
		{
			name:         "plan with mutations",
			addMutations: true,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := NewPlan()
			if tt.addMutations {
				plan.Add(spanner.Insert("table", []string{"col"}, []interface{}{"val"}))
			}

			assert.Equal(t, tt.expected, plan.IsEmpty())
		})
	}
}

func TestPlan_Count(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		numMutations  int
		expectedCount int
	}{
		{
			name:          "empty plan",
			numMutations:  0,
			expectedCount: 0,
		},
		{
			name:          "plan with one mutation",
			numMutations:  1,
			expectedCount: 1,
		},
		{
			name:          "plan with multiple mutations",
			numMutations:  5,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := NewPlan()
			for i := 0; i < tt.numMutations; i++ {
				plan.Add(spanner.Insert("table", []string{"col"}, []interface{}{"val"}))
			}

			assert.Equal(t, tt.expectedCount, plan.Count())
		})
	}
}

func TestPlan_Clear(t *testing.T) {
	t.Parallel()

	plan := NewPlan()
	plan.Add(spanner.Insert("table1", []string{"col"}, []interface{}{"val1"}))
	plan.Add(spanner.Insert("table2", []string{"col"}, []interface{}{"val2"}))

	assert.False(t, plan.IsEmpty())

	plan.Clear()

	assert.True(t, plan.IsEmpty())
	assert.Empty(t, plan.Mutations())
}

func TestNewCommitter(t *testing.T) {
	t.Parallel()

	committer := NewCommitter(nil)
	assert.NotNil(t, committer)
}

func TestCommitter_Client(t *testing.T) {
	t.Parallel()

	committer := NewCommitter(nil)
	assert.Nil(t, committer.Client())
}
