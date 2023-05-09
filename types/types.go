package types

type TestId string
type BootstrapPeers []string
type PrivateRegistry string

type Shutdown chan struct{}

func PtrString(str string) *string {
	return &str
}

func GetString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
