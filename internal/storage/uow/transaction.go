package uow

import (
	"github.com/jackc/pgx/v5"
)

type Transaction struct {
	repositories map[RepoName]RepositoryFactory
	tx           pgx.Tx
}

func NewTransaction(tx pgx.Tx, repositories map[RepoName]RepositoryFactory) *Transaction {
	return &Transaction{
		repositories: repositories,
		tx:           tx,
	}
}

// Get возвращает репозиторий или ошибку ErrRepositoryNotRegistered.
func (t *Transaction) Get(name RepoName) (Repository, error) {
	if repo, ok := t.repositories[name]; ok {
		return repo(t.tx), nil
	}
	return nil, ErrRepositoryNotRegistered
}

// GetAs возвращает зарегистрированный репозиторий с именем name приведенный к типу T
// или ошибки ErrRepositoryNotRegistered в случае не найденного репозитория с указанным name, ErrInvalidRepositoryType

func GetAs[T any](t TX, name RepoName) (T, error) {
	repo, err := t.Get(name)
	var res T
	if err != nil {
		return res, err //nolint:wrapcheck
	}
	res, ok := repo.(T)
	if !ok {
		return res, ErrInvalidRepositoryType
	}
	return res, nil
}
