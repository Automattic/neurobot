package bot

import (
	"neurobot/infrastructure/matrix"
)

type Registry interface {
}

type registry struct {
	homeserverURL string
	clients       map[string]matrix.Client
}

func NewRegistry(homeserverURL string) *registry {
	return &registry{homeserverURL: homeserverURL}
}
