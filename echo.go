// echo is a simple HTTP server.
package main

//go:generate go run lesiw.io/plain/cmd/plaingen@latest

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/netip"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"labs.lesiw.io/ctr"
	"labs.lesiw.io/echo/internal/stmt"
	"lesiw.io/defers"
	"lesiw.io/plain"
)

func main() {
	defer defers.Run()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		defers.Exit(1)
	}
}

var db *pgxpool.Pool
var ctx = context.Background()

// ez execs f if v is zero.
func ez[T comparable](v T, f func() error) error {
	var zero T
	if v != zero {
		return nil
	}
	return f()
}

var hostname string

var hits = expvar.NewInt("hits")

func run() error {
	if err := ez(os.Getenv("PGHOST"), ctr.Postgres); err != nil {
		return fmt.Errorf("failed to set up postgres: %w", err)
	}
	db = plain.ConnectPgx(ctx)
	defers.Add(db.Close)

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "<unknown>"
		slog.Error(err.Error())
	}

	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	l, err := net.Listen("tcp4", ":8080")
	if err != nil {
		return fmt.Errorf("could not listen on port 8080: %w", err)
	}
	return http.Serve(l, echo)
}

var echo http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	hits.Add(1)
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	for k, v := range r.Header {
		if k == "X-Real-Ip" && len(v) > 0 && v[0] != "" {
			ip = v[0]
			break
		}
	}
	slog.Info("http request",
		"method", r.Method,
		"path", r.URL.Path,
		"address", ip,
		"query", r.URL.RawQuery,
	)

	now := time.Now()
	go logRequest(ctx, now, r, ip)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := writeBody(w, r, now); err != nil {
		slog.Error("writeBody() err", "error", err)
	}
}

func writeBody(
	w http.ResponseWriter, r *http.Request, now time.Time,
) (err error) {
	tw := tabwriter.NewWriter(w, 0, 8, 1, '\t', tabwriter.AlignRight)
	write := func(format string, a ...any) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(tw, format, a...)
	}
	flush := func() {
		if err != nil {
			return
		}
		err = tw.Flush()
	}

	write("Hostname: %s\n\n", hostname)
	flush()

	write("Request Information:\n")
	write("\tTimestamp:\t%s\n", now)
	write("\tPath:\t%s\n", r.URL.Path)
	write("\tMethod:\t%s\n", r.Method)
	write("\tAddress:\t%s\n", r.RemoteAddr)
	write("\tQuery:\t%s\n", r.URL.RawQuery)
	write("\n")
	flush()

	write("Headers:\n")
	for _, h := range slices.Sorted(maps.Keys(r.Header)) {
		write("\t%s\t%s\n", h, strings.Join(r.Header[h], ","))
	}
	flush()

	return
}

func logRequest(
	ctx context.Context, now time.Time, r *http.Request, ipstr string,
) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("could not read request body", "error", err.Error())
		body = []byte("")
	}

	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = strings.Join(values, ",")
	}
	headersJSON, err := json.Marshal(headers)
	if err != nil {
		slog.Error("could not marshal headers", "error", err.Error())
		headersJSON = []byte("")
	}

	ip, err := netip.ParseAddr(ipstr)
	if err != nil {
		slog.Error("could not parse ip", "error", err.Error())
	}

	_, err = db.Exec(ctx, stmt.AddRequest,
		now,
		r.URL.String(),
		r.Method,
		ip,
		headersJSON,
		string(body),
	)
	if err != nil {
		slog.Error("could not add request to db", "error", err.Error())
	}
}
