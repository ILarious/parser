package service

import (
	"context"
	"errors"
	"testing"

	"parser/internal/domain/model"
	parserkafka "parser/internal/infrastructure/kafka"
)

func TestParserServiceProcessOrderPublishesProcessingAndDone(t *testing.T) {
	repo := &fakeRepository{claimed: true}
	accounts := &fakeAccountClient{account: model.VKAccount{
		SocialID:       1,
		AccountType:    "profile",
		FullName:       "Pavel Durov",
		Username:       "durov",
		FollowersCount: 100,
	}}
	producer := &fakeProducer{}
	svc := NewParserService(repo, accounts, producer)

	err := svc.ProcessOrder(context.Background(), "1", "orders", parserkafka.OrderRequest{
		EventID:  1,
		OrderID:  10,
		Username: "durov",
	})
	if err != nil {
		t.Fatalf("ProcessOrder() error = %v", err)
	}

	if !repo.completed {
		t.Fatal("expected order completion")
	}
	if len(producer.messages) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(producer.messages))
	}
	if producer.messages[0].Status != int(model.OrderStatusProcessing) {
		t.Fatalf("expected processing status, got %d", producer.messages[0].Status)
	}
	if producer.messages[1].Status != int(model.OrderStatusDone) {
		t.Fatalf("expected done status, got %d", producer.messages[1].Status)
	}
	if producer.messages[1].FullName != "Pavel Durov" || producer.messages[1].FollowersCount != 100 {
		t.Fatalf("unexpected done payload: %+v", producer.messages[1])
	}
}

func TestParserServiceProcessOrderPublishesFailedOnAccountError(t *testing.T) {
	repo := &fakeRepository{claimed: true}
	accounts := &fakeAccountClient{err: errors.New("not found")}
	producer := &fakeProducer{}
	svc := NewParserService(repo, accounts, producer)

	err := svc.ProcessOrder(context.Background(), "1", "orders", parserkafka.OrderRequest{
		EventID:  1,
		OrderID:  10,
		Username: "missing",
	})
	if err != nil {
		t.Fatalf("ProcessOrder() error = %v", err)
	}

	if !repo.failed {
		t.Fatal("expected order failure")
	}
	if len(producer.messages) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(producer.messages))
	}
	if producer.messages[1].Status != int(model.OrderStatusFailed) {
		t.Fatalf("expected failed status, got %d", producer.messages[1].Status)
	}
}

func TestParserServiceProcessOrderSkipsAlreadyProcessedMessage(t *testing.T) {
	repo := &fakeRepository{claimed: false}
	accounts := &fakeAccountClient{}
	producer := &fakeProducer{}
	svc := NewParserService(repo, accounts, producer)

	err := svc.ProcessOrder(context.Background(), "1", "orders", parserkafka.OrderRequest{
		EventID:  1,
		OrderID:  10,
		Username: "durov",
	})
	if err != nil {
		t.Fatalf("ProcessOrder() error = %v", err)
	}
	if accounts.called {
		t.Fatal("account client should not be called")
	}
	if len(producer.messages) != 0 {
		t.Fatalf("expected no published messages, got %d", len(producer.messages))
	}
}

type fakeRepository struct {
	claimed   bool
	completed bool
	failed    bool
}

func (r *fakeRepository) ClaimOrder(context.Context, string, string, int64, int64, string) (bool, error) {
	return r.claimed, nil
}

func (r *fakeRepository) CompleteOrder(_ context.Context, _ int64, account model.VKAccount) (model.VKAccount, error) {
	r.completed = true
	account.ID = 1
	return account, nil
}

func (r *fakeRepository) FailOrder(context.Context, int64, string) error {
	r.failed = true
	return nil
}

type fakeAccountClient struct {
	account model.VKAccount
	err     error
	called  bool
}

func (c *fakeAccountClient) GetAccountInfo(context.Context, string) (model.VKAccount, error) {
	c.called = true
	if c.err != nil {
		return model.VKAccount{}, c.err
	}
	return c.account, nil
}

type fakeProducer struct {
	messages []parserkafka.OrderResponse
}

func (p *fakeProducer) Publish(_ context.Context, response parserkafka.OrderResponse) error {
	p.messages = append(p.messages, response)
	return nil
}
