# TinyCloud

TinyCloud is a TUI for turning a small Dockerized project into a shareable app.

The first milestone is intentionally local:

1. Run `tinycloud`.
2. Add a project directory containing a `Dockerfile`.
3. Deploy it with Docker.
4. Open the generated `http://<name>.localhost` URL.
5. Inspect logs or stop the app from the same TUI.

This validates the core workflow before adding a hosted control plane, HTTPS,
authentication, or Kubernetes.

## Run

```sh
go run ./cmd/tinycloud
```

Docker must be installed and available on `PATH`. Use `n` to add an app, `d`
to deploy the selected app, `l` to view logs, `s` to stop it, and `q` to quit.

