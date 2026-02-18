// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	projectID  = "test-project"
	instanceID = "test-instance"
	databaseID = "test-database"
)

func main() {
	ctx := context.Background()

	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost == "" {
		emulatorHost = "localhost:9010"
		os.Setenv("SPANNER_EMULATOR_HOST", emulatorHost)
	}

	fmt.Printf("Using Spanner emulator at: %s\n", emulatorHost)

	// Create gRPC connection options for emulator
	opts := []option.ClientOption{
		option.WithEndpoint(emulatorHost),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
	}

	// Create instance
	if err := createInstance(ctx, opts); err != nil {
		log.Printf("Instance creation: %v (may already exist)", err)
	}

	// Create database
	if err := createDatabase(ctx, opts); err != nil {
		log.Printf("Database creation: %v (may already exist)", err)
	}

	fmt.Println("Setup complete!")
}

func createInstance(ctx context.Context, opts []option.ClientOption) error {
	instanceAdmin, err := instance.NewInstanceAdminClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create instance admin client: %v", err)
	}
	defer instanceAdmin.Close()

	op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectID),
		InstanceId: instanceID,
		Instance: &instancepb.Instance{
			Config:      fmt.Sprintf("projects/%s/instanceConfigs/emulator-config", projectID),
			DisplayName: "Test Instance",
			NodeCount:   1,
		},
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Instance created successfully")
	return nil
}

func createDatabase(ctx context.Context, opts []option.ClientOption) error {
	databaseAdmin, err := database.NewDatabaseAdminClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create database admin client: %v", err)
	}
	defer databaseAdmin.Close()

	op, err := databaseAdmin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseID),
		ExtraStatements: []string{
			`CREATE TABLE products (
				product_id STRING(36) NOT NULL,
				name STRING(255) NOT NULL,
				description STRING(MAX),
				category STRING(100) NOT NULL,
				base_price_numerator INT64 NOT NULL,
				base_price_denominator INT64 NOT NULL,
				discount_percent NUMERIC,
				discount_start_date TIMESTAMP,
				discount_end_date TIMESTAMP,
				status STRING(20) NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				archived_at TIMESTAMP,
			) PRIMARY KEY (product_id)`,
			`CREATE TABLE outbox_events (
				event_id STRING(36) NOT NULL,
				event_type STRING(100) NOT NULL,
				aggregate_id STRING(36) NOT NULL,
				payload JSON NOT NULL,
				status STRING(20) NOT NULL,
				created_at TIMESTAMP NOT NULL,
				processed_at TIMESTAMP,
			) PRIMARY KEY (event_id)`,
			`CREATE INDEX idx_outbox_status ON outbox_events(status, created_at)`,
			`CREATE INDEX idx_products_category ON products(category, status)`,
		},
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Database created successfully")
	return nil
}
