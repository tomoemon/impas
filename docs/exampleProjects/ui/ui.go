package ui

import (
	"fmt"
	"github.com/tomoemon/impas/docs/exampleProjects/infra"
)

//noinspection GoUnusedExportedFunction
func PrintUser(id string) {
	repo := infra.UserRepoImpl{}
	user := repo.Get(id)
	fmt.Printf("user: %+v", user)
}
