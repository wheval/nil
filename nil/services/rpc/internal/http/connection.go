package http

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/version"
)

const (
	MaxRequestContentLength  = 1024 * 1024 * 32 // 32MB
	minSupportedNiljsVersion = "0.24.0"
)

var minSupportedRevision = 1

type (
	remoteCtxKey    struct{}
	schemeCtxKey    struct{}
	localCtxKey     struct{}
	userAgentCtxKey struct{}
	originCtxKey    struct{}
)

// HttpServerConn turns a HTTP connection into a Conn.
type HttpServerConn struct {
	io.Reader
	io.Writer
	Request *http.Request
}

func init() {
	num, err := strconv.Atoi(version.GetGitRevision())
	check.PanicIfErr(err)
	minSupportedRevision = num
}

// Close does nothing and always returns nil.
func (t *HttpServerConn) Close() error { return nil }

// RemoteAddr returns the peer address of the underlying connection.
func (t *HttpServerConn) RemoteAddr() string {
	return t.Request.RemoteAddr
}

// SetWriteDeadline does nothing and always returns nil.
func (t *HttpServerConn) SetWriteDeadline(time.Time) error { return nil }
