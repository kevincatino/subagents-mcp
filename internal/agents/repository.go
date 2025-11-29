package agents

import "context"

// Repository provides agent discovery.
type Repository interface {
	ListAgents(ctx context.Context) ([]Agent, error)
}
