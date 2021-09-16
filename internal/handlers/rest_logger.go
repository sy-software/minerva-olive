package handlers

import (
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sy-software/minerva-olive/internal/core/domain"
)

type ServerCtxKeys string

// LogValues represents the values we want to include in server logs
type LogValues struct {
	ReqId      string
	SerName    string
	Path       string
	Latency    time.Duration
	Method     string
	StatusCode int
	ClientIP   string
	ForwardFor []string
	MsgStr     string
	Body       interface{}
}

// logValuesFromCtx values from the context ready to be logged
func logValuesFromCtx(serName string, c *gin.Context) *LogValues {
	t := time.Now()

	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	if raw != "" {
		path = path + "?" + raw
	}
	msg := c.Errors.String()
	if msg == "" {
		msg = "Request"
	}

	reqId := c.Request.Context().Value(domain.RequestIdKey)
	reqIdStr := ""
	if reqId != nil {
		reqIdStr = reqId.(string)
	}

	body, _ := ioutil.ReadAll(c.Request.Body)
	bodyStr := ""
	if body != nil {
		bodyStr = string(body)
	}

	return &LogValues{
		Path:       path,
		Method:     c.Request.Method,
		ReqId:      reqIdStr,
		SerName:    serName,
		Latency:    time.Since(t),
		StatusCode: c.Writer.Status(),
		ClientIP:   c.ClientIP(),
		ForwardFor: c.Request.Header.Values("X-Forwarded-For"),
		MsgStr:     msg,
		Body:       bodyStr,
	}
}

// logSwitch prints LogValues into the right log stream using zerolog
func logSwitch(data *LogValues) {
	var logger *zerolog.Event
	switch {
	case data.StatusCode >= 400 && data.StatusCode < 500:
		logger = log.Warn()
	case data.StatusCode >= 500:
		logger = log.Error()
	default:
		logger = log.Info()
	}

	logger.
		Str("req_path", data.Path).
		Str("req_method", data.Method).
		Str("req_id", data.ReqId).
		Interface("req_body", data.Body).
		Str("ser_name", data.SerName).
		Dur("resp_time", data.Latency).
		Int("status", data.StatusCode).
		Str("client_ip", data.ClientIP).
		Strs("fwd_for", data.ForwardFor).
		Msg(data.MsgStr)
}

// LogMiddleware is a Gin server logger middleware
func LogMiddleware(serName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logSwitch(logValuesFromCtx(serName, c))
	}
}
