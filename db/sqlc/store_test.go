package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	/// run a concurrent transfer transaction
	n := 5

	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransaferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	/// check results

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results

		require.NotEmpty(t, result)

		//// check for transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//	check entries

		fromEntity := result.FromEntry
		require.NotEmpty(t, fromEntity)
		require.Equal(t, account1.ID, fromEntity.AccountID)
		require.Equal(t, -amount, fromEntity.Amount)
		require.NotZero(t, fromEntity.ID)
		require.NotZero(t, fromEntity.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntity.ID)
		require.NoError(t, err)

		//	check entries

		toEntity := result.ToEntry
		require.NotEmpty(t, toEntity)
		require.Equal(t, account2.ID, toEntity.AccountID)
		require.Equal(t, amount, toEntity.Amount)
		require.NotZero(t, toEntity.ID)
		require.NotZero(t, toEntity.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntity.ID)
		require.NoError(t, err)

		/// TODO: check account balance
	}
}
