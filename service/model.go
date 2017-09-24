package service

import "github.com/docker/docker/api/types"

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type Errorer interface {
	error() error
}

//type CreateNetworksResponse struct {
//	Networks []types.NetworkResource `json:"networks,omitempty"`
//	Err      error                   `json:"err,omitempty"`
//}

//func (r CreateNetworksResponse) error() error { return r.Err }

type CreateNetworkResponse struct {
	Network types.NetworkCreateResponse `json:"network,omitempty"`
	Err     error                       `json:"err,omitempty"`
}

func (r CreateNetworkResponse) error() error { return r.Err }

type NetworkConf struct {
	NetworkScope string
	SubnetMask   int
	Blacklist    []string
}

type Conf struct {
	Overlay NetworkConf
	Bridge  NetworkConf
}
