package oui

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gookit/gcli/v3/progress"
)

type Options struct {
	Logger                 *LoggerType
	Progress               *progress.Progress
	Version                string
	Connection             *sql.DB
	dialect                int
	MaxConnections         uint
	HTTPClient             *http.Client
	InlineRowsPerStatement int
}

type Option func(*Options)

func sanitizeVersion(v string) string {
	if len(v) == 0 {
		return "default"
	}
	first := regexp.MustCompile(`^[0-9].*`)
	if first.MatchString(v) {
		v = fmt.Sprintf("oui_%s", v)
	}
	repl := regexp.MustCompile(`[^A-Za-z0-9_]`)
	v = repl.ReplaceAllString(v, "__")
	return v
}

func WithProgress(p *progress.Progress) Option {
	return func(opts *Options) {
		opts.Progress = p
	}
}

func WithLogging(logger LoggerType) Option {
	return func(opts *Options) {
		opts.Logger = &logger
	}
}

func WithVersion(version string) Option {
	return func(opts *Options) {
		opts.Version = sanitizeVersion(version)
	}
}

func WithConnection(conn *sql.DB) Option {
	return func(opts *Options) {
		opts.Connection = conn
	}
}

func WithMaxConnections(max uint) Option {
	return func(opts *Options) {
		opts.MaxConnections = max
	}
}

// WithSQLiteConnection uses an existing connection that speaks the SQLite
// dialect, such as a Cloudflare D1 binding, instead of opening one via
// CreateSQLiteOption.
func WithSQLiteConnection(conn *sql.DB) Option {
	return func(opts *Options) {
		opts.Connection = conn
		opts.dialect = dialectSqlite
	}
}

// WithInlineBulkInsert makes BulkInsert render values as escaped SQL literals
// instead of bound parameters, chunked at maxRowsPerStatement rows. Use this
// with backends that cap bound parameters or queries per request (e.g.
// Cloudflare D1: 100 parameters/statement, 1,000 queries/invocation).
// A maxRowsPerStatement <= 0 selects the default of 250.
func WithInlineBulkInsert(maxRowsPerStatement int) Option {
	return func(opts *Options) {
		if maxRowsPerStatement <= 0 {
			maxRowsPerStatement = defaultInlineRows
		}
		opts.InlineRowsPerStatement = maxRowsPerStatement
	}
}

// WithHTTPClient sets the client used to download registry CSVs during
// Populate, for environments where the default client is unavailable
// (e.g. fetch-backed clients on Cloudflare Workers).
func WithHTTPClient(client *http.Client) Option {
	return func(opts *Options) {
		opts.HTTPClient = client
	}
}

func getOptions(setters ...Option) *Options {
	options := &Options{
		Logger:         nil,
		Progress:       nil,
		Version:        "default",
		Connection:     nil,
		MaxConnections: 0,
	}
	for _, setter := range setters {
		setter(options)
	}
	return options
}
