package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"parser/internal/domain/model"
	parserkafka "parser/internal/infrastructure/kafka"
)

type ParsingOrderRepository interface {
	ClaimOrder(ctx context.Context, messageID, topic string, eventID, orderID int64, username string) (bool, error)
	CompleteOrder(ctx context.Context, orderID int64, account model.VKAccount) (model.VKAccount, error)
	FailOrder(ctx context.Context, orderID int64, reason string) error
}

type AccountClient interface {
	GetAccountInfo(ctx context.Context, username string) (model.VKAccount, error)
}

type ResultProducer interface {
	Publish(ctx context.Context, response parserkafka.OrderResponse) error
}

type ParserService struct {
	orders   ParsingOrderRepository
	accounts AccountClient
	producer ResultProducer
}

func NewParserService(orders ParsingOrderRepository, accounts AccountClient, producer ResultProducer) *ParserService {
	return &ParserService{
		orders:   orders,
		accounts: accounts,
		producer: producer,
	}
}

func (s *ParserService) ProcessOrder(ctx context.Context, messageID, topic string, request parserkafka.OrderRequest) error {
	request.Username = strings.TrimSpace(request.Username)
	if request.OrderID == 0 || request.Username == "" {
		return fmt.Errorf("invalid order request: order_id and username are required")
	}

	claimed, err := s.orders.ClaimOrder(ctx, messageID, topic, request.EventID, request.OrderID, request.Username)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	if err := s.publishStatus(ctx, request, model.OrderStatusProcessing, model.VKAccount{}); err != nil {
		return err
	}

	account, err := s.accounts.GetAccountInfo(ctx, request.Username)
	if err != nil {
		reason := err.Error()
		if failErr := s.orders.FailOrder(ctx, request.OrderID, reason); failErr != nil {
			return failErr
		}
		if publishErr := s.publishStatus(ctx, request, model.OrderStatusFailed, model.VKAccount{}); publishErr != nil {
			return publishErr
		}
		log.Printf("failed to parse vk account %q: %v", request.Username, err)
		return nil
	}

	savedAccount, err := s.orders.CompleteOrder(ctx, request.OrderID, account)
	if err != nil {
		return err
	}

	return s.publishStatus(ctx, request, model.OrderStatusDone, savedAccount)
}

func (s *ParserService) publishStatus(ctx context.Context, request parserkafka.OrderRequest, status model.OrderStatus, account model.VKAccount) error {
	return s.producer.Publish(ctx, parserkafka.OrderResponse{
		EventID:        parserkafka.ResponseEventID(request.OrderID, int(status)),
		OrderID:        request.OrderID,
		Username:       request.Username,
		FullName:       account.FullName,
		FollowersCount: account.FollowersCount,
		Status:         int(status),
	})
}
