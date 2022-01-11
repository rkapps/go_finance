package core

import (
	"context"
	"time"

	"github.com/rkapps/go_finance/store"
)

func (fn *Finance) AggregateTransactions(ctx context.Context, fromDate *time.Time, toDate *time.Time) ([]store.TransactionAgg, error) {
	return fn.MDB.AggregateTransactions(ctx, fromDate, toDate)
}
