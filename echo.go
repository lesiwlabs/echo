// echo is a simple HTTP server.
package main

//go:generate go run lesiw.io/plain/cmd/plaingen@latest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/netip"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

func run() error {
	db = plain.ConnectPgx(ctx)
	defers.Add(db.Close)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "<unknown>"
		slog.Error(err.Error())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		go logRequest(ctx, now, r)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "Hostname: %s\n\n", hostname)

		fmt.Fprintf(w, "Request Information:\n")
		tw := tabwriter.NewWriter(w, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintf(tw, "\tTimestamp:\t%s\n", now)
		fmt.Fprintf(tw, "\tPath:\t%s\n", r.URL.Path)
		fmt.Fprintf(tw, "\tMethod:\t%s\n", r.Method)
		fmt.Fprintf(tw, "\tAddress:\t%s\n", r.RemoteAddr)
		fmt.Fprintf(tw, "\tQuery:\t%s\n", r.URL.RawQuery)
		fmt.Fprintf(tw, "\n")
		tw.Flush()

		fmt.Fprintf(w, "Headers:\n")
		for _, h := range slices.Sorted(maps.Keys(r.Header)) {
			fmt.Fprintf(tw, "\t%s\t%s\n", h, strings.Join(r.Header[h], ","))
		}
		tw.Flush()
	})

	return http.ListenAndServe(":8080", nil)
}

func logRequest(ctx context.Context, now time.Time, r *http.Request) {
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

	var ip netip.Addr
	ipPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		slog.Error("could not parse ip", "error", err.Error())
	} else {
		ip = ipPort.Addr()
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
