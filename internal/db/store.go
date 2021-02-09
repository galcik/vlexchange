package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/galcik/vlexchange/internal/currency"
	"github.com/galcik/vlexchange/internal/db/queries"
	_ "github.com/lib/pq"
)

type Store struct {
	db      *sql.DB
	queries *queries.Queries
	context context.Context
}

func NewStore(connectionString string) (*Store, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Store{db: db, queries: queries.New(db)}, nil
}

func (store *Store) WithContext(ctx context.Context) *Store {
	return &Store{db: store.db, queries: store.queries, context: ctx}
}

func (store *Store) ExecuteTx(transaction func(context.Context, *queries.Queries) error) error {
	ctx := store.context
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := queries.New(tx)
	err = transaction(store.context, q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (store *Store) GetAccountByToken(token string) (*queries.Account, error) {
	var account queries.Account
	var err error
	err = store.ExecuteTx(
		func(ctx context.Context, q *queries.Queries) error {
			account, err = q.GetAccountByToken(ctx, token)
			return err
		},
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (store *Store) GetAccount(accountId int32) (*queries.Account, error) {
	var account queries.Account
	var err error
	err = store.ExecuteTx(
		func(ctx context.Context, q *queries.Queries) error {
			account, err = q.GetAccountById(ctx, accountId)
			return err
		},
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (store *Store) GetStandingOrder(orderId int32) (*queries.StandingOrder, error) {
	var order queries.StandingOrder
	var err error
	err = store.ExecuteTx(
		func(ctx context.Context, q *queries.Queries) error {
			order, err = q.GetStandingOrder(ctx, orderId)
			return err
		},
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (store *Store) GetStandingOrders(orderIds []int32) ([]queries.StandingOrder, error) {
	var err error
	var orders []queries.StandingOrder
	err = store.ExecuteTx(
		func(ctx context.Context, q *queries.Queries) error {
			orders, err = q.GetStandingOrders(ctx, orderIds)
			return err
		},
	)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

type CreateStandingOrderParams struct {
	AccountID  int32
	OrderType  queries.OrderType
	Quantity   currency.BTC
	LimitPrice currency.USD
}

func (store *Store) CreateStandingOrder(params CreateStandingOrderParams) (
	*queries.StandingOrder,
	[]int32,
	error,
) {
	reservedUSD := currency.USD(0)
	reservedBTC := currency.BTC(0)
	if params.OrderType == queries.OrderTypeBuy {
		reservedUSD = params.Quantity.USD(params.LimitPrice.Float64())
	} else {
		reservedBTC = params.Quantity
	}

	var standingOrder queries.StandingOrder
	var affectedOrderIds []int32
	err := store.ExecuteTx(
		func(ctx context.Context, q *queries.Queries) error {
			account, err := q.GetAccountById(ctx, params.AccountID)
			if err != nil {
				return err
			}

			reservedAmounts, err := q.GetReservedAmounts(ctx, params.AccountID)
			if err != nil {
				return err
			}

			sufficientAmounts := (reservedAmounts.UsdAmount+reservedUSD.Internal() <= account.UsdAmount) &&
				(reservedAmounts.BtcAmount+reservedBTC.Internal() <= account.BtcAmount)
			state := queries.OrderStateLive
			if !sufficientAmounts {
				state = queries.OrderStateCancelled
				reservedBTC = 0
				reservedUSD = 0
			}
			standingOrder, err = q.CreateStandingOrder(
				ctx,
				queries.CreateStandingOrderParams{
					AccountID:         params.AccountID,
					Type:              params.OrderType,
					State:             state,
					Quantity:          params.Quantity.Internal(),
					LimitPrice:        params.LimitPrice.Internal(),
					ReservedBtcAmount: reservedBTC.Internal(),
					ReservedUsdAmount: reservedUSD.Internal(),
				},
			)

			if err != nil {
				return err
			}

			affectedOrderIds = append(affectedOrderIds, standingOrder.ID)

			if state != queries.OrderStateLive {
				return nil
			}

			for standingOrder.State != queries.OrderStateFulfilled {
				if params.OrderType == queries.OrderTypeBuy {
					sellOrder, err := q.GetBestSeller(ctx, params.LimitPrice.Internal())
					if errors.Is(err, sql.ErrNoRows) {
						return nil
					}

					affectedOrderIds = append(affectedOrderIds, sellOrder.ID)
					if err := processDeal(ctx, q, &sellOrder, &standingOrder, sellOrder.LimitPrice); err != nil {
						return err
					}
				} else {
					buyOrder, err := q.GetBestBuyer(ctx, params.LimitPrice.Internal())
					if errors.Is(err, sql.ErrNoRows) {
						return nil
					}

					affectedOrderIds = append(affectedOrderIds, buyOrder.ID)
					if err := processDeal(ctx, q, &standingOrder, &buyOrder, buyOrder.LimitPrice); err != nil {
						return err
					}
				}
			}

			return nil
		},
	)

	return &standingOrder, affectedOrderIds, err
}

func processDeal(
	ctx context.Context,
	q *queries.Queries,
	sellOrder *queries.StandingOrder,
	buyOrder *queries.StandingOrder,
	btcPrice int64,
) error {
	quantity := sellOrder.Quantity
	if buyOrder.Quantity < quantity {
		quantity = buyOrder.Quantity
	}

	dealPrice := currency.BTC(quantity).USD(currency.USD(btcPrice).Float64()).Internal()
	updatedRows, err := q.TransferAmounts(
		ctx,
		queries.TransferAmountsParams{ID: sellOrder.AccountID, UsdAmount: dealPrice, BtcAmount: -quantity},
	)
	if err != nil {
		return err
	}
	if updatedRows != 1 {
		return fmt.Errorf("invalid transfer from seller")
	}

	updatedRows, err = q.TransferAmounts(
		ctx,
		queries.TransferAmountsParams{ID: buyOrder.AccountID, UsdAmount: -dealPrice, BtcAmount: quantity},
	)
	if err != nil {
		return err
	}
	if updatedRows != 1 {
		return fmt.Errorf("invalid transfer to buy")
	}

	reservedBtcChange := quantity
	*sellOrder, err = q.SatisfyOrder(
		ctx,
		queries.SatisfyOrderParams{
			ID:                sellOrder.ID,
			Quantity:          quantity,
			FilledPrice:       dealPrice,
			ReservedBtcAmount: reservedBtcChange,
		},
	)
	if err != nil {
		return err
	}

	reservedUsdChange := currency.BTC(quantity).USD(currency.USD(buyOrder.LimitPrice).Float64()).Internal()
	*buyOrder, err = q.SatisfyOrder(
		ctx,
		queries.SatisfyOrderParams{
			ID:                buyOrder.ID,
			Quantity:          quantity,
			FilledPrice:       dealPrice,
			ReservedUsdAmount: reservedUsdChange,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (store *Store) DeleteStandingOrder(orderId int32) error {
	return store.ExecuteTx(
		func(ctx context.Context, q *queries.Queries) error {
			return q.DeleteStandingOrder(ctx, orderId)
		},
	)
}
