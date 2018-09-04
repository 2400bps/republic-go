package contract

// Network is used to represent a Republic Protocol network.
type Network string

const (
	// NetworkMainnet represents the mainnet
	NetworkMainnet Network = "mainnet"
	// NetworkTestnet represents the internal F∅ testnet
	NetworkTestnet Network = "testnet"
	// NetworkFalcon represents the internal Falcon testnet
	NetworkFalcon Network = "falcon"
	// NetworkNightly represents the internal Nightly testnet
	NetworkNightly Network = "nightly"
	// NetworkLocal represents a local network
	NetworkLocal Network = "local"
)

// Config defines the different settings for connecting to Ethereum on
// different Republic Protocol networks.
type Config struct {
	Network                    Network `json:"network"`
	URI                        string  `json:"uri"`
	RepublicTokenAddress       string  `json:"republicTokenAddress"`
	DarknodeRegistryAddress    string  `json:"darknodeRegistryAddress"`
	DarknodeRewardVaultAddress string  `json:"darknodeRewardVaultAddress"`
	DarknodeSlasherAddress     string  `json:"darkodeSlasherAddress"`
	OrderbookAddress           string  `json:"orderbookAddress"`
	SettlementRegistryAddress  string  `json:"settlementRegistryAddress"`
}

// IsNil returns true if Config or any of its fields are nil.
func (config *Config) IsNil() bool {
	if config == nil || len(config.Network) == 0 || len(config.URI) == 0 || len(config.RepublicTokenAddress) == 0 || len(config.DarknodeRegistryAddress) == 0 || len(config.OrderbookAddress) == 0 {
		return true
	}
	return false
}
