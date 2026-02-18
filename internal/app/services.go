package app

import (
	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/app/product/queries"
	"github.com/product-catalog-service/internal/app/product/repo"
	"github.com/product-catalog-service/internal/app/product/usecases"
	"github.com/product-catalog-service/internal/clock"
	"github.com/product-catalog-service/internal/committer"
	grpchandler "github.com/product-catalog-service/internal/transport/grpc/product"
)

// Services holds all application services and dependencies.
type Services struct {
	// Infrastructure
	SpannerClient *spanner.Client
	Committer     *committer.Committer
	Clock         clock.Clock

	// Repositories
	ProductRepo *repo.ProductRepo
	OutboxRepo  *repo.OutboxRepo
	ReadModel   *repo.ProductReadModel

	// Use Cases (Commands)
	UseCases *usecases.ProductUseCases

	// Queries
	Queries *queries.ProductQueries

	// gRPC Handler
	ProductHandler *grpchandler.Handler
}

// NewServices creates and wires all application services.
func NewServices(spannerClient *spanner.Client) *Services {
	s := &Services{}

	// Infrastructure
	s.SpannerClient = spannerClient
	s.Committer = committer.NewCommitter(spannerClient)
	s.Clock = clock.NewRealClock()

	// Repositories
	s.ProductRepo = repo.NewProductRepo(spannerClient)
	s.OutboxRepo = repo.NewOutboxRepo()
	s.ReadModel = repo.NewProductReadModel(spannerClient)

	// Use Cases (Commands) - consolidated
	s.UseCases = usecases.NewProductUseCases(s.ProductRepo, s.OutboxRepo, s.Committer, s.Clock)

	// Queries - consolidated
	s.Queries = queries.NewProductQueries(s.ReadModel, s.Clock)

	// gRPC Handler
	s.ProductHandler = grpchandler.NewHandler(s.UseCases, s.Queries)

	return s
}

// Close cleans up resources.
func (s *Services) Close() {
	if s.SpannerClient != nil {
		s.SpannerClient.Close()
	}
}
