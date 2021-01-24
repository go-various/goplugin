package logical

import (
	"context"
	"fmt"
	"github.com/go-various/consul"
	"github.com/hashicorp/consul/api"
	"time"
)

var _ Consul = (*ConsulView)(nil)

const SessionPath = "session"

type ConsulView struct {
	consul    consul.Client
}

func (c *ConsulView) Config(ctx context.Context) (consul.Config, error) {
	return *c.consul.Config(), nil
}

func NewConsulView(consul consul.Client) *ConsulView {
	return &ConsulView{
		consul:    consul,
	}
}

func (c *ConsulView) GetServiceAddrPort(ctx context.Context, name string, useLan bool, tags string) (host string, port int, err error) {
	return c.consul.GetServiceAddrPort(name, useLan, tags)
}

func (c *ConsulView) GetService(ctx context.Context, id, tag string) (*api.AgentService, error) {
	return c.consul.GetService(id, tag)
}

func (c *ConsulView) NewSession(ctx context.Context,
	name string, ttl time.Duration, behavior consul.SessionBehavior) (string, error) {
	return c.consul.NewSession(name, ttl, behavior, &api.WriteOptions{})
}

func (c *ConsulView) SessionInfo(ctx context.Context, id string) (*api.SessionEntry, error) {
	return c.consul.SessionInfo(id, &api.QueryOptions{})
}

func (c *ConsulView) DestroySession(ctx context.Context, id string) error {
	return c.consul.DestroySession(id, &api.WriteOptions{})
}

func (c *ConsulView) KVAcquire(ctx context.Context, key, session string) (success bool, err error) {
	return c.consul.KVAcquire(c.expendSessionKey(key), session, &api.QueryOptions{})
}

func (c *ConsulView) KVRelease(ctx context.Context, key string) error {
	return c.consul.KVRelease(c.expendSessionKey(key), &api.QueryOptions{})
}

func (c *ConsulView) KVInfo(ctx context.Context, key string) (*api.KVPair, error) {
	return c.consul.KVInfo(key, &api.QueryOptions{})
}

func (c *ConsulView) KVCas(ctx context.Context, p *api.KVPair) (bool, error) {
	return c.consul.KVCas(p, &api.WriteOptions{})
}

func (c *ConsulView) KVList(ctx context.Context, prefix string) (api.KVPairs, error) {
	kvps, _, err := c.consul.Client().KV().List(prefix, &api.QueryOptions{})
	return kvps, err
}

func (c *ConsulView) KVCreate(ctx context.Context, p *api.KVPair) error {
	_, err := c.consul.Client().KV().Put(p, &api.WriteOptions{})
	return err
}

func (c *ConsulView) expendSessionKey(key string) string {
	return fmt.Sprintf("/%s/%s", SessionPath, key)
}
