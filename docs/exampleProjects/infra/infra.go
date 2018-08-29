package infra

import "github.com/tomoemon/impas/docs/exampleProjects/domain"

type UserRepoImpl struct {
}

func (*UserRepoImpl) Get(id string) *domain.User {
	return &domain.User{
		ID:   "id1",
		Name: "name1",
	}
}
