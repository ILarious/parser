package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"parser/config"
	"parser/internal/domain/service"
	parserkafka "parser/internal/infrastructure/kafka"
	parserpostgres "parser/internal/infrastructure/postgres"
	"parser/internal/infrastructure/postgres/migration"
	"parser/internal/infrastructure/vkontakte"
	"parser/pkg/postgres"
	"parser/pkg/worker_pool"
)

func main() {
	cfg := config.Load()

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer func() {
		if err := postgres.Close(db); err != nil {
			log.Printf("failed to close postgres: %v", err)
		}
	}()

	migrationCtx, cancelMigration := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelMigration()
	if err := migration.Up(migrationCtx, db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	orderRepository, err := parserpostgres.NewParsingOrderRepository(db)
	if err != nil {
		log.Fatalf("failed to create parsing order repository: %v", err)
	}

	workers, err := worker_pool.NewWorkerPool(cfg.WorkerPool.Size)
	if err != nil {
		log.Fatalf("failed to create worker pool: %v", err)
	}

	vkClient := vkontakte.NewClient(cfg.VK.BaseURL, cfg.VK.APIVersion, cfg.VK.AccessToken, cfg.VK.Timeout)
	orderProducer := parserkafka.NewOrderProducer(cfg.Kafka.Brokers, cfg.Kafka.OrderResponseTopic)
	parserService := service.NewParserService(orderRepository, vkClient, orderProducer)
	orderConsumer := parserkafka.NewOrderConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.OrderRequestTopic,
		cfg.Kafka.OrderRequestGroupID,
		parserService,
		workers,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go orderConsumer.Run(ctx)

	<-ctx.Done()
	stop()

	if err := orderConsumer.Close(); err != nil {
		log.Printf("failed to close order consumer: %v", err)
	}

	workers.Close()
	workers.Wait()

	if err := orderProducer.Close(); err != nil {
		log.Printf("failed to close order producer: %v", err)
	}
}
