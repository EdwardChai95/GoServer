package ws

import "time"

type EthAddressArray struct {
	List []string
}

//EAA for eth address array
var EAA EthAddressArray

type EthAddressTimeArray struct {
	List []time.Time
}

//EAT for eth address time array
var EAT EthAddressTimeArray
