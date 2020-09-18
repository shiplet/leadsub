# `leadsub`

To build
```bash
$ go build
```

To run
```bash
$ go run leadsub

$ go build
$ ./leadsub
```

You'll need to populate `Ids` in a `data/data.go` file, and provide the necessary data.

```go
const Ids = []string{"randomLeadUUID", "randomLeadID"}
```

You'll also need to create three values in `env/env.go`:

```go
const PHP_SESSID = ""
const API_KEY = ""
const API_SECRET = ""
```