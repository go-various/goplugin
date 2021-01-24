package logical

import (
	"context"
	"github.com/go-various/consul"
	"github.com/hashicorp/consul/api"
	"time"
)

type Consul interface {
	Config(ctx context.Context) (consul.Config, error)

	//获取服务列表
	GetService(ctx context.Context, id, tag string) (*api.AgentService, error)

	//获取微服务路径
	GetServiceAddrPort(ctx context.Context, name string, useLan bool, tags string) (host string, port int, err error)

	//创建一个session,ttl需大于15秒,behavior定义了session到期后的动作
	//如需深度定制session请获取native客户端创建
	NewSession(ctx context.Context, name string, ttl time.Duration, behavior consul.SessionBehavior) (string, error)
	SessionInfo(ctx context.Context, id string) (*api.SessionEntry, error)
	//销毁session
	DestroySession(ctx context.Context, id string) error

	//对一个kv进行加锁
	//err==nil && success 加锁成功
	KVAcquire(ctx context.Context, key, session string) (success bool, err error)
	//释放一个session的锁
	KVRelease(ctx context.Context, key string) error
	//获取kv信息
	KVInfo(ctx context.Context, key string) (*api.KVPair, error)
	//检查或者设置key
	KVCas(ctx context.Context, p *api.KVPair) (bool, error)
	KVList(ctx context.Context, prefix string) (api.KVPairs, error)
	KVCreate(ctx context.Context, p *api.KVPair) error
}
