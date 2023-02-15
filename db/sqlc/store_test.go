package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	existed := make(map[int]bool)
	fmt.Println(">>>> Before: ", account1.Balance, account2.Balance)
	/// run a concurrent transfer transaction
	n := 5

	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransaferTxResult)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
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
		//// check accounts

		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, account2.ID)

		/// TODO: check account balance
		fmt.Println(">>>>tx: ", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) /// 1 * amount, 2 * amount, ... n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		existed[k] = true
	}
	///check the final updated balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	fmt.Println(">>>> After: ", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">>>> Before: ", account1.Balance, account2.Balance)
	/// run a concurrent transfer transaction
	n := 10

	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID
		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err

		}()
	}

	/// check results

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

	}
	///check the final updated balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	fmt.Println(">>>> After: ", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
