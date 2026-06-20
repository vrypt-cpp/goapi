package health

import (
	"net/http"
	"time"

	"goapi/core"
)

type Plugin struct {
	startedAt time.Time
}

func New() *Plugin {
	return &Plugin{startedAt: time.Now()}
}

func (p *Plugin) Name() string {
	return "health"
}

type StatusResponse struct {
	Status    string `json:"status"`
	UptimeSec int64  `json:"uptime_sec"`
}

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{
			Method:   "GET",
			Path:     "/health",
			Summary:  "Health check",
			Tags:     []string{"health"},
			Response: StatusResponse{},
			Handler: func(w http.ResponseWriter, r *http.Request) {
				core.JSON(w, http.StatusOK, StatusResponse{
					Status:    "ok",
					UptimeSec: int64(time.Since(p.startedAt).Seconds()),
				})
			},
		},
	}
}
