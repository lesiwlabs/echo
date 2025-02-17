package main

import (
	"os"

	"labs.lesiw.io/ops/goapp"
	k8sapp "labs.lesiw.io/ops/k8s/goapp"
	"lesiw.io/ops"
)

type Ops struct{ k8sapp.Ops }

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "build")
	}
	goapp.Name = "echo"
	o := Ops{}
	o.Postgres = true
	o.Hostname = "echo.lesiw.dev"
	ops.Handle(o)
}
