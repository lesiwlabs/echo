package main

import (
	"os"

	"labs.lesiw.io/ops/github"
	"labs.lesiw.io/ops/goapp"
	k8sapp "labs.lesiw.io/ops/k8s/goapp"
	"lesiw.io/ops"
)

type k8sOps = k8sapp.Ops
type ghOps = github.Ops

type Ops struct {
	k8sOps
	ghOps
}

var secrets = map[string]string{
	"TEST_SECRET": "testsecret",
}

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "build")
	}
	github.Repo = "lesiwlabs/echo"
	github.Secrets = secrets
	goapp.Name = "echo"
	o := Ops{}
	o.EnvSecrets = secrets
	o.Hostname = "echo.lesiw.dev"
	o.Postgres = true
	o.Port = 8080
	ops.Handle(o)
}
