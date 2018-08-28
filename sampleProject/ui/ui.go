package ui

import (
	"fmt"
	"github.com/tomoemon/assert-dep/sampleProject/infra"
)

func PrintUser(id string) {
	repo := infra.UserRepoImpl{}
	user := repo.Get(id)
	fmt.Printf("user: %+v", user)
}
