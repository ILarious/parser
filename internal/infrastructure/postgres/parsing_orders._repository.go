package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"parser/internal/domain/model"
)

var ErrNilDB = errors.New("parsing order repository: nil db")

type ParsingOrderRepository struct {
	db *sql.DB
}

func NewParsingOrderRepository(db *sql.DB) (*ParsingOrderRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	return &ParsingOrderRepository{db: db}, nil
}

func (r *ParsingOrderRepository) ClaimOrder(ctx context.Context, messageID, topic string, eventID, orderID int64, username string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin claim order: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_ = messageID
	_ = topic

	result, err := tx.ExecContext(ctx, `
		INSERT INTO parsing_orders (event_id, order_id, username, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (order_id) DO UPDATE
		SET event_id = EXCLUDED.event_id,
			username = EXCLUDED.username,
			status = EXCLUDED.status,
			error_text = '',
			updated_at = NOW()
		WHERE parsing_orders.status <> $5
	`, eventID, orderID, username, int(model.ParsingStatusProcessing), int(model.ParsingStatusDone))
	if err != nil {
		return false, fmt.Errorf("create parsing order: %w", err)
	}

	changed, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check claimed parsing order: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit claim order: %w", err)
	}

	return changed > 0, nil
}

func (r *ParsingOrderRepository) CompleteOrder(ctx context.Context, orderID int64, account model.VKAccount) (model.VKAccount, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.VKAccount{}, fmt.Errorf("begin complete order: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	savedAccount, err := upsertAccount(ctx, tx, account)
	if err != nil {
		return model.VKAccount{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE parsing_orders
		SET status = $2,
			error_text = '',
			account_id = $3,
			updated_at = NOW()
		WHERE order_id = $1
	`, orderID, int(model.ParsingStatusDone), savedAccount.ID); err != nil {
		return model.VKAccount{}, fmt.Errorf("mark parsing order done: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.VKAccount{}, fmt.Errorf("commit complete order: %w", err)
	}

	return savedAccount, nil
}

func (r *ParsingOrderRepository) FailOrder(ctx context.Context, orderID int64, reason string) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE parsing_orders
		SET status = $2,
			error_text = $3,
			updated_at = NOW()
		WHERE order_id = $1
	`, orderID, int(model.ParsingStatusFailed), reason); err != nil {
		return fmt.Errorf("mark parsing order failed: %w", err)
	}

	return nil
}

func upsertAccount(ctx context.Context, tx *sql.Tx, account model.VKAccount) (model.VKAccount, error) {
	const query = `
		INSERT INTO vk_accounts (
			social_id,
			account_type,
			full_name,
			username,
			followers_count,
			avatar_url,
			private,
			verified
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (account_type, social_id) DO UPDATE
		SET full_name = EXCLUDED.full_name,
			username = EXCLUDED.username,
			followers_count = EXCLUDED.followers_count,
			avatar_url = EXCLUDED.avatar_url,
			private = EXCLUDED.private,
			verified = EXCLUDED.verified,
			updated_at = NOW()
		RETURNING id, social_id, account_type, full_name, username, followers_count, avatar_url, private, verified, created_at, updated_at
	`

	var saved model.VKAccount
	if err := tx.QueryRowContext(
		ctx,
		query,
		account.SocialID,
		account.AccountType,
		account.FullName,
		account.Username,
		account.FollowersCount,
		account.AvatarURL,
		account.Private,
		account.Verified,
	).Scan(
		&saved.ID,
		&saved.SocialID,
		&saved.AccountType,
		&saved.FullName,
		&saved.Username,
		&saved.FollowersCount,
		&saved.AvatarURL,
		&saved.Private,
		&saved.Verified,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	); err != nil {
		return model.VKAccount{}, fmt.Errorf("upsert vk account: %w", err)
	}

	return saved, nil
}
