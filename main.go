// Copyright (c) bwplotka/mimic Authors
// Licensed under the Apache License 2.0.

package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/google/uuid"
	"github.com/hunjixin/mytest/env"
	"github.com/hunjixin/mytest/utils"
	"k8s.io/apimachinery/pkg/util/json"
)

func main() {
	fmt.Println(runVenusAuth(context.Background()))
}

func runVenusAuth(ctx context.Context) error {
	controller, err := env.NewEnvController("default", uuid.New().String(), true)
	if err != nil {
		return err
	}

	client, closer, err := controller.RunVenusAuth(ctx, "script/simple")
	if err != nil {
		return err
	}

	defer closer()

	_, err = client.CreateUser(&auth.CreateUserRequest{
		Name:    "li",
		Comment: utils.StringPtr("comment li"),
		State:   0,
	})
	if err != nil {
		return err
	}

	users, err := client.ListUsers(&auth.ListUsersRequest{
		Page: &core.Page{
			Skip:  0,
			Limit: 0,
		},
	})
	if err != nil {
		return err
	}

	usersBytes, err := json.Marshal(users)
	if err != nil {
		return err
	}
	fmt.Printf("%v", string(usersBytes))
	utils.Prompt()
	return nil
}
