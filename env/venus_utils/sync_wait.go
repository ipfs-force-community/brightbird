package venusutils

import (
	"context"
	"strconv"
	"time"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"
)

var log = logging.Logger("env")

// SyncWait returns when ChainHead is within 20 epochs of the expected height
func SyncWait(ctx context.Context, k8sEnv *env.K8sEnvDeployer, pod corev1.Pod, port int, adminToken string) error {
	endpoint := types.EndpointFromString(pod.Status.PodIP + ":" + strconv.Itoa(port))
	napi, closer, err := v1.DialFullNodeRPC(ctx, endpoint.ToMultiAddr(), adminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	params, err := napi.StateGetNetworkParams(ctx)
	if err != nil {
		return err
	}
	for {
		state, err := napi.SyncState(ctx)
		if err != nil {
			return err
		}

		if len(state.ActiveSyncs) == 0 {
			time.Sleep(time.Second * 2)
			continue
		}

		head, err := napi.ChainHead(ctx)
		if err != nil {
			return err
		}

		working := -1
		for i, ss := range state.ActiveSyncs {
			switch ss.Stage {
			case vTypes.StageSyncComplete:
			case vTypes.StageIdle:
				// not complete, not actively working
			default:
				working = i
			}
		}

		if working == -1 {
			working = len(state.ActiveSyncs) - 1
		}

		ss := state.ActiveSyncs[working]

		if ss.Base == nil || ss.Target == nil {
			log.Info(
				"syncing",
				"height", ss.Height,
				"stage", ss.Stage.String(),
			)
		} else {
			log.Info(
				"syncing",
				"base", ss.Base.Key(),
				"target", ss.Target.Key(),
				"target_height", ss.Target.Height(),
				"height", ss.Height,
				"stage", ss.Stage.String(),
			)
		}

		if time.Now().Unix()-int64(head.MinTimestamp()) < int64(params.BlockDelaySecs*2) {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(int64(params.BlockDelaySecs) * int64(time.Second))):
		}
	}

	return nil
}
