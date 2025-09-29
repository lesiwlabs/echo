package main

import (
	"testing"

	errname "github.com/Antonboom/errname/pkg/analyzer"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/gofix"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"lesiw.io/checker"
	"lesiw.io/errcheck/errcheck"
	"lesiw.io/linelen"
	"lesiw.io/tidytypes"
)

func TestCheck(t *testing.T) {
	checker.Run(t,
		atomicalign.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errcheck.Analyzer,
		errname.New(),
		fieldalignment.Analyzer,
		gofix.Analyzer,
		hostport.Analyzer,
		httpmux.Analyzer,
		linelen.Analyzer,
		nilness.Analyzer,
		reflectvaluecompare.Analyzer,
		sortslice.Analyzer,
		tidytypes.Analyzer,
		unusedwrite.Analyzer,
		waitgroup.Analyzer,
	)
}
