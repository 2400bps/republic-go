package adapter

import "github.com/republicprotocol/republic-go/status"

// Status defines a structure for JSON marshalling
type Status struct {
	Network         string `json:"network"`
	MultiAddress    string `json:"multiAddress"`
	EthereumAddress string `json:"ethereumAddress"`
	Peers           int    `json:"peers"`
}

// StatusAdapter defines a struct which has status reading capability
type StatusAdapter struct {
	status.Reader
}

// NewStatusAdapter returns an adapter which contains reading statuses
func NewStatusAdapter(reader status.Reader) StatusAdapter {
	return StatusAdapter{
		Reader: reader,
	}
}

// Status returns a Status object with populated fields
func (adapter *StatusAdapter) Status() (Status, error) {
	network, err := adapter.Network()
	if err != nil {
		return Status{}, err
	}
	multiAddr, err := adapter.MultiAddress()
	if err != nil {
		return Status{}, err
	}
	ethAddr, err := adapter.EthereumAddress()
	if err != nil {
		return Status{}, err
	}
	peers, err := adapter.Peers()
	if err != nil {
		return Status{}, err
	}
	return Status{
		Network:         network,
		MultiAddress:    multiAddr.String(),
		EthereumAddress: ethAddr,
		Peers:           peers,
	}, nil
}
