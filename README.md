# echo

This is an example lesiwlabs service. It should be used as an example of best
practices for more complex services, and will be kept up-to-date as those 
practices evolve.

## Local development

A container runtime (e.g. Docker, podman, containerd) is required.

The application will start up its own PostgreSQL container unless it is
configured to point to an existing PostgreSQL instance using
[libpq environment variables][libpq]. So running the application locally should
be as simple as `go run .`

## Known issues

* GitHub Actions lacks adequate permissions to build and deploy.
* No scheduled backups.

[libpq]: https://www.postgresql.org/docs/current/libpq-envars.html
