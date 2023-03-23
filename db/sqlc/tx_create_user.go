package db

import "context"

//// CreateUserTxParams contains the input parameters of the transfer transaction

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

// TransaferTxResult is the result of the transfer transaction

type CreateUserTxResult struct {
	User User
}

/// Transfer performs a money transfer from one account to the other
/// it create a transfer record, add account entries, and update accounts

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		err = arg.AfterCreate(result.User)
		return err
	})
	return result, err
}
