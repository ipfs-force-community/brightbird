package types

type AdminToken string
type TestId string
type BootstrapPeers []string
type PrivateRegistry string

type Shutdown chan struct{}

func PtrString(str string) *string {
	return &str
}
