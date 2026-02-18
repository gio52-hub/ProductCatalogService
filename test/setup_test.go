package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/query"
	"github.com/product-catalog-service/internal/repository"
	"github.com/product-catalog-service/internal/usecase"
	"github.com/product-catalog-service/internal/clock"
	"github.com/product-catalog-service/internal/committer"
)

const (
	testProject  = "test-project"
	testInstance = "test-instance"
	testDatabase = "test-database"
)

// TestFixture holds all test dependencies.
type TestFixture struct {
	ctx           context.Context
	spannerClient *spanner.Client
	committer     *committer.Committer
	clock         *clock.FixedClock

	// Repositories
	ProductRepo *repository.ProductRepo
	OutboxRepo  *repository.OutboxRepo
	ReadModel   *repository.ProductReadModel

	// Use Cases
	UseCases *usecase.ProductUseCases

	// Queries
	Queries *query.ProductQueries
}

// SetupTestFixture creates a new test fixture with all dependencies.
func SetupTestFixture(t *testing.T) *TestFixture {
	t.Helper()

	ctx := context.Background()

	// Check if SPANNER_EMULATOR_HOST is set
	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost == "" {
		t.Skip("SPANNER_EMULATOR_HOST not set, skipping E2E tests")
	}

	// Build database path
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", testProject, testInstance, testDatabase)

	// Create Spanner client
	spannerClient, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to create Spanner client: %v", err)
	}

	// Use a fixed clock for deterministic tests
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fixedClock := clock.NewFixedClock(fixedTime)

	// Initialize infrastructure
	comm := committer.NewCommitter(spannerClient)

	// Repositories
	productRepo := repository.NewProductRepo(spannerClient)
	outboxRepo := repository.NewOutboxRepo()
	readModel := repository.NewProductReadModel(spannerClient)

	fixture := &TestFixture{
		ctx:           ctx,
		spannerClient: spannerClient,
		committer:     comm,
		clock:         fixedClock,

		ProductRepo: productRepo,
		OutboxRepo:  outboxRepo,
		ReadModel:   readModel,

		// Use Cases (consolidated)
		UseCases: usecase.NewProductUseCases(productRepo, outboxRepo, comm, fixedClock),

		// Queries (consolidated)
		Queries: query.NewProductQueries(readModel, fixedClock),
	}

	t.Cleanup(func() {
		fixture.spannerClient.Close()
	})

	return fixture
}

// AdvanceTime advances the fixture's clock by the given duration.
func (f *TestFixture) AdvanceTime(d time.Duration) {
	f.clock.Advance(d)
}

// SetTime sets the fixture's clock to the given time.
func (f *TestFixture) SetTime(t time.Time) {
	f.clock.SetTime(t)
}

// Now returns the current time from the fixture's clock.
func (f *TestFixture) Now() time.Time {
	return f.clock.Now()
}

// Context returns the test context.
func (f *TestFixture) Context() context.Context {
	return f.ctx
}

// GetOutboxEvents retrieves outbox events for a given aggregate ID.
func (f *TestFixture) GetOutboxEvents(t *testing.T, aggregateID string) []OutboxEventRow {
	t.Helper()

	stmt := spanner.Statement{
		SQL: `SELECT event_id, event_type, aggregate_id, status, created_at 
		      FROM outbox_events 
		      WHERE aggregate_id = @aggregate_id 
		      ORDER BY created_at`,
		Params: map[string]interface{}{
			"aggregate_id": aggregateID,
		},
	}

	iter := f.spannerClient.Single().Query(f.ctx, stmt)
	defer iter.Stop()

	var events []OutboxEventRow
	for {
		row, err := iter.Next()
		if err != nil {
			break
		}

		var event OutboxEventRow
		if err := row.Columns(&event.EventID, &event.EventType, &event.AggregateID, &event.Status, &event.CreatedAt); err != nil {
			t.Fatalf("Failed to read outbox event: %v", err)
		}
		events = append(events, event)
	}

	return events
}

// OutboxEventRow represents a row from the outbox_events table.
type OutboxEventRow struct {
	EventID     string
	EventType   string
	AggregateID string
	Status      string
	CreatedAt   time.Time
}

// CleanupProduct deletes a product by ID (for test cleanup).
func (f *TestFixture) CleanupProduct(t *testing.T, productID string) {
	t.Helper()

	mut := spanner.Delete("products", spanner.Key{productID})
	_, err := f.spannerClient.Apply(f.ctx, []*spanner.Mutation{mut})
	if err != nil {
		t.Logf("Warning: failed to cleanup product %s: %v", productID, err)
	}

	// Also cleanup outbox events
	stmt := spanner.Statement{
		SQL: `SELECT event_id FROM outbox_events WHERE aggregate_id = @aggregate_id`,
		Params: map[string]interface{}{
			"aggregate_id": productID,
		},
	}
	iter := f.spannerClient.Single().Query(f.ctx, stmt)
	defer iter.Stop()

	var keys []spanner.Key
	for {
		row, err := iter.Next()
		if err != nil {
			break
		}
		var eventID string
		row.Columns(&eventID)
		keys = append(keys, spanner.Key{eventID})
	}

	if len(keys) > 0 {
		keySet := spanner.KeySetFromKeys(keys...)
		mut := spanner.Delete("outbox_events", keySet)
		f.spannerClient.Apply(f.ctx, []*spanner.Mutation{mut})
	}
}
