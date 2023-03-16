package env

const (
	VenusDaemon        DeployType = "venus"
	ChainCo            DeployType = "chain-co"
	MarketClient       DeployType = "market-client"
	VenusAuth          DeployType = "venus-auth"
	VenusGateway       DeployType = "venus-gateway"
	VenusMarket        DeployType = "venus-market"
	VenusMessage       DeployType = "venus-message"
	VenusMiner         DeployType = "venus-miner"
	VenusWallet        DeployType = "venus-wallet"
	VenusSectorManager DeployType = "venus-sector-manager"
	VenusWorker        DeployType = "venus-worker"
)

type DeployType string
