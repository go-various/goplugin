package logical

import (
	"context"
	"fmt"
	"github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/metric"
	"github.com/go-various/goplugin/pluginregister"
	"github.com/go-various/helper/jsonutil"
	"time"
)

type WorkerData struct {
	Backend string
	Request *logical.Request
}

//网关处理客户端http请求
func (m *Transport) backend() func(i interface{}) (interface{}, error) {
	return func(i interface{}) (result interface{}, err error) {
		data := i.(*WorkerData)
		backendName := data.Backend
		metric.PluginCountMetric.
			WithLabelValues(backendName , data.Request.Namespace, data.Request.Operation).Add(1)

		defer func(then time.Time) {
			if nil != err {
				m.Logger.Error("backend", "id", data.Request.ID,
					"name", backendName, "namespace", data.Request.Namespace,
					"operation", data.Request.Operation,
					"status", "finished",
					"Err", err, "took", time.Since(then))
			} else {
				if m.Logger.IsTrace() {
					m.Logger.Trace("backend", "id", data.Request.ID,
						"name", backendName, "namespace", data.Request.Namespace,
						"operation", data.Request.Operation,
						"status", "finished",
						"took", time.Since(then))
				}
			}
			metric.PluginGaugeMetric.
				WithLabelValues(backendName , data.Request.Namespace, data.Request.Operation).
				Set(time.Since(then).Seconds())

		}(time.Now())

		if m.Logger.IsTrace() {
			m.Logger.Trace("backend", "id", data.Request.ID, "name", backendName,
				"namespace", data.Request.Namespace, "operation", data.Request.Operation,
				"status", "started", "request", jsonutil.EncodeToString(data.Request))
		}

		backend, has := m.PluginManager.GetBackend(backendName)
		if !has {
			return nil, pluginregister.PluginNotExists
		}

		backend.Incr()
		defer backend.DeIncr()
		if m.authEnabled && m.authMethod != nil {
			authReply, err := m.authorization(backend, data.Request)
			if err != nil {
				return nil, fmt.Errorf("auth: %s", err.Error())
			}
			if authReply.ResultCode != 0 {
				return authReply, nil
			}
			if err := jsonutil.Swap(authReply.Content.Data, &data.Request.Authorized); err != nil {
				return nil, err
			}
		}

		return backend.HandleRequest(context.Background(), data.Request)
	}
}
