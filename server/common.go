package server

import (
	"github.com/trafficstars/fasthttp"
	"github.com/trafficstars/statuspage"
)

func writeMetrics(ctx *fasthttp.RequestCtx) {
	statuspage.WriteMetricsPrometheus(ctx.Response.BodyWriter())
}
