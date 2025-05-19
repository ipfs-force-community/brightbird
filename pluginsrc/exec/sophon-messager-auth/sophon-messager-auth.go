package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	mTypes "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/env/plugin"
	"github.com/ipfs-force-community/brightbird/pkg/types"
	"github.com/ipfs-force-community/brightbird/pkg/version"

	sophonmessager "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-messager"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "sophon-messager-auth-check",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "sophon-messager auth token check",
}

type TestCaseParams struct {
	Messager   sophonmessager.SophonMessagerReturn `json:"SophonMessager"  jsonschema:"SophonMessager"  title:"Sophon Messager" require:"true" description:"messager return"`
	Token      string                              `json:"token" jsonschema:"token" title:"token" require:"true" description:"token to check"`
	Permission string                              `json:"permission" jsonschema:"permission" title:"Permission" require:"true" description:"read,write,admin,sign"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {

	switch params.Permission {
	case "read", "write", "admin", "sign":
		break
	default:
		return fmt.Errorf("unknown permission %s", params.Permission)
	}

	permission, err := checkPermission(ctx, params)
	if err != nil {
		return err
	}
	if !strings.Contains(permission, params.Permission) {
		return fmt.Errorf("token:%v api-check:[%v], not pass", params.Permission, permission)
	}
	fmt.Printf("token:%v api-check:[%v] , pass\n", params.Permission, permission)

	return err
}

func checkPermission(ctx context.Context, params TestCaseParams) (string, error) {
	client, closer, err := messager.DialIMessagerRPC(ctx, params.Messager.SvcEndpoint.ToMultiAddr(), params.Token, nil)
	if err != nil {
		return "", err
	}
	defer closer()

	read, write, sign, admin := false, false, false, false

	// check read
	if _, err = client.Version(ctx); err == nil {
		read = true
	}

	// check write
	if _, err := client.ClearUnFillMessage(ctx, address.Address{}); err == nil {
		write = true
	}
	if write {
		// check sign
		// TODO: check if this is a valid way to check sign permission
		fmt.Println("check sign permission not support yet")
		if _, err := client.Send(ctx, mTypes.QuickSendParams{}); err == nil {
			sign = true
		}
		// check admin
		if _, err = client.NetPeers(ctx); err == nil {
			admin = true
		}
	}

	resStr := ""
	if read {
		resStr += "read "
	}
	if write {
		resStr += "write "
	}
	if sign {
		resStr += "sign "
	}
	if admin {
		resStr += "admin "
	}

	return resStr, nil
}
