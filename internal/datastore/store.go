package datastore

//go:generate mockery -all

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/galcik/vlexchange/internal/currency"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	_ "github.com/lib/pq"
)

type Store interface {
	WithContext(ctx context.Context) Store
	ExecuteTx(transaction func(context.Context, queries.Querier) error) error

	GetAccountByToken(token string) (*queries.Account, error)
	GetAccount(accountId int32) (*queries.Account, error)
	DepositAccount(accountId int32, btcAmount currency.BTC, usdAmount currency.USD) (bool, error)

	ExecuteMarketOrder(params CreateMarketOrderParams) (
		CreateMarketOrderResult,
		[]int32,
		error,
	)

	CreateStandingOrder(params CreateStandingOrderParams) (
		*queries.StandingOrder,
		[]int32,
		error,
	)
	GetStandingOrder(orderId int32) (*queries.StandingOrder, error)
	GetStandingOrders(orderIds []int32) ([]queries.StandingOrder, error)
	DeleteStandingOrder(orderId int32) error
}

type DbStore struct {
	db      *sql.DB
	querier queries.Querier
	context context.Context
}

func NewStore(db *sql.DB) (Store, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DbStore{
		querier: queries.New(db),
		db:      db,
		context: context.Background(),
	}, nil
}

func (store *DbStore) WithContext(ctx context.Context) Store {
	return &DbStore{db: store.db, querier: store.querier, context: ctx}
}

func (store *DbStore) ExecuteTx(transaction func(context.Context, queries.Querier) error) error {
	ctx := store.context
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
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

func (store *DbStore) GetAccountByToken(token string) (*queries.Account, error) {
	var account queries.Account
	var err error
	err = store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
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

func (store *DbStore) GetAccount(accountId int32) (*queries.Account, error) {
	var account queries.Account
	var err error
	err = store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
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

func (store *DbStore) DepositAccount(accountId int32, btcAmount currency.BTC, usdAmount currency.USD) (bool, error) {
	var err error
	var success bool
	err = store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
			rowCount, err := q.TransferAmounts(
				ctx,
				queries.TransferAmountsParams{
					ID:        accountId,
					UsdAmount: usdAmount.Internal(),
					BtcAmount: btcAmount.Internal(),
				},
			)
			success = rowCount == 1
			return err
		},
	)
	return success, err
}

func (store *DbStore) GetStandingOrder(orderId int32) (*queries.StandingOrder, error) {
	var order queries.StandingOrder
	var err error
	err = store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
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

func (store *DbStore) GetStandingOrders(orderIds []int32) ([]queries.StandingOrder, error) {
	var err error
	var orders []queries.StandingOrder
	err = store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
			orders, err = q.GetStandingOrders(ctx, orderIds)
			return err
		},
	)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

type CreateMarketOrderParams struct {
	AccountID int32
	OrderType queries.OrderType
	Quantity  currency.BTC
}

type CreateMarketOrderResult struct {
	Quantity currency.BTC
	Price    currency.USD
}

func (store *DbStore) ExecuteMarketOrder(params CreateMarketOrderParams) (
	CreateMarketOrderResult,
	[]int32,
	error,
) {
	var result CreateMarketOrderResult
	var affectedOrderIds []int32
	err := store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
			standingOrder, err := q.CreateStandingOrder(
				ctx,
				queries.CreateStandingOrderParams{
					AccountID:         params.AccountID,
					Type:              params.OrderType,
					State:             queries.OrderStateLive,
					Quantity:          params.Quantity.Internal(),
					LimitPrice:        0,
					ReservedBtcAmount: 0,
					ReservedUsdAmount: 0,
				},
			)
			if err != nil {
				return err
			}

			for standingOrder.Quantity > 0 {
				var account queries.Account
				account, err = q.GetAccountById(ctx, params.AccountID)
				if err != nil {
					return err
				}

				if params.OrderType == queries.OrderTypeBuy {
					sellOrder, err := q.GetBestMarketSeller(ctx)
					if errors.Is(err, sql.ErrNoRows) {
						break
					}

					affectedOrderIds = append(affectedOrderIds, sellOrder.ID)
					btcPrice := sellOrder.LimitPrice
					maxBuyQuantity := currency.NewBTC(float64(account.UsdAmount) / float64(btcPrice)).Internal()
					quantity := minQuantity(sellOrder.Quantity, maxBuyQuantity, standingOrder.Quantity)
					if err := processDeal(ctx, q, &sellOrder, &standingOrder, quantity, btcPrice); err != nil {
						return err
					}
				} else {
					buyOrder, err := q.GetBestMarketBuyer(ctx)
					if errors.Is(err, sql.ErrNoRows) {
						break
					}

					affectedOrderIds = append(affectedOrderIds, buyOrder.ID)
					btcPrice := buyOrder.LimitPrice
					quantity := minQuantity(standingOrder.Quantity, buyOrder.Quantity)
					if err := processDeal(ctx, q, &standingOrder, &buyOrder, quantity, btcPrice); err != nil {
						return err
					}
				}
			}

			result.Quantity = currency.BTC(standingOrder.FilledQuantity)
			result.Price = currency.USD(standingOrder.FilledPrice)

			err = q.DeleteStandingOrder(ctx, standingOrder.ID)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return result, affectedOrderIds, err
}

type CreateStandingOrderParams struct {
	AccountID  int32
	OrderType  queries.OrderType
	Quantity   currency.BTC
	LimitPrice currency.USD
}

func (store *DbStore) CreateStandingOrder(params CreateStandingOrderParams) (
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
		func(ctx context.Context, q queries.Querier) error {
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
					quantity := minQuantity(standingOrder.Quantity, sellOrder.Quantity)
					btcPrice := sellOrder.LimitPrice
					if err := processDeal(ctx, q, &sellOrder, &standingOrder, quantity, btcPrice); err != nil {
						return err
					}
				} else {
					buyOrder, err := q.GetBestBuyer(ctx, params.LimitPrice.Internal())
					if errors.Is(err, sql.ErrNoRows) {
						return nil
					}

					affectedOrderIds = append(affectedOrderIds, buyOrder.ID)
					quantity := minQuantity(standingOrder.Quantity, buyOrder.Quantity)
					btcPrice := buyOrder.LimitPrice
					if err := processDeal(ctx, q, &standingOrder, &buyOrder, quantity, btcPrice); err != nil {
						return err
					}
				}
			}

			return nil
		},
	)

	return &standingOrder, affectedOrderIds, err
}

type dealProcessingResult struct {
	USDAmount currency.USD
}

func processDeal(
	ctx context.Context,
	q queries.Querier,
	sellOrder *queries.StandingOrder,
	buyOrder *queries.StandingOrder,
	quantity int64,
	btcPrice int64,
) error {
	var result dealProcessingResult
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

	result.USDAmount = currency.USD(dealPrice)
	return nil
}

func (store *DbStore) DeleteStandingOrder(orderId int32) error {
	return store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
			return q.DeleteStandingOrder(ctx, orderId)
		},
	)
}

func minQuantity(amounts ...int64) int64 {
	result := amounts[0]
	for _, amount := range amounts {
		if amount < result {
			result = amount
		}
	}
	return result
}
