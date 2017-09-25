package service

import "errors"

var (
	ErrNoMoreNetworks = errors.New("No avaliable networks.")
	ErrNotInSwarmMode = errors.New("No running Swarm cluster.")
)
