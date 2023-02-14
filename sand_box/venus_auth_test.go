package main_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/google/uuid"
	"github.com/hunjixin/mytest/env"
	"github.com/hunjixin/mytest/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVenusAuth(t *testing.T) {
	timeout := flag.Int64("timeout", 0, "timeout for running test")
	ctx := context.Background()
	if timeout != nil && *timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Second*time.Duration(*timeout))
	}

	controller, err := env.NewEnvController("default", uuid.New().String(), true)
	assert.Nil(t, err)

	client, closer, err := controller.RunVenusAuth(ctx, "../script/simple")
	assert.Nil(t, err)

	defer closer()

	_, err = client.CreateUser(&auth.CreateUserRequest{
		Name:    "li",
		Comment: utils.StringPtr("comment li"),
		State:   0,
	})
	assert.Nil(t, err)

	users, err := client.ListUsers(&auth.ListUsersRequest{
		Page: &core.Page{
			Skip:  0,
			Limit: 0,
		},
	})
	assert.Nil(t, err)

	usersBytes, err := json.Marshal(users)
	assert.Nil(t, err)
	fmt.Printf("%v", string(usersBytes))

}
