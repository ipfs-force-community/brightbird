package venusutils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"
	corev1 "k8s.io/api/core/v1"

	"github.com/ipfs-force-community/brightbird/pkg/env"
	"github.com/ipfs-force-community/brightbird/pkg/types"
)

var venusLog = logging.Logger("env")

// SyncWait returns when ChainHead is within 20 epochs of the expected height
func SyncWait(ctx context.Context, k8sEnv *env.K8sEnvDeployer, pod corev1.Pod, port int, adminToken string) error {
	endpoint := types.EndpointFromString(pod.Status.PodIP + ":" + strconv.Itoa(port))
	napi, closer, err := v1.DialFullNodeRPC(ctx, endpoint.ToMultiAddr(), adminToken, nil)
	if err != nil {
		return err
	}
	defer closer()

	currHeight, maxHeight := abi.ChainEpoch(-1), abi.ChainEpoch(-1)
	var params *vTypes.NetworkParams
	for params == nil {
		if params, err = napi.StateGetNetworkParams(ctx); err == nil && params != nil {
			break
		}
		// max wait 3 blocks
		if head, err := napi.ChainHead(ctx); err == nil && head != nil {
			if head.Height() != currHeight {
				venusLog.Infof("chain head %v:%s", head.Height(), head.String())
			}
			if currHeight == -1 || maxHeight == -1 {
				currHeight = head.Height()
				maxHeight = currHeight + 3
			}
			if head.Height() > maxHeight {
				return fmt.Errorf("get network params failed after wait 3 block height")
			}
		} else {
			return err
		}
		time.Sleep(time.Second * 3)
	}
	if params == nil {
		return fmt.Errorf("get network params failed")
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
			venusLog.With(
				"base", ss.Base.Key(),
				"target", ss.Target.Key(),
			).Info("syncing")
		} else {
			venusLog.With(
				"base", ss.Base.Key(),
				"target", ss.Target.Key(),
				"target_height", ss.Target.Height(),
				"height", ss.Height,
				"stage", ss.Stage.String(),
			).Info("syncing")
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
