# raft-grpc-leader-rpc

[![Godoc](https://godoc.org/github.com/Jille/raft-grpc-leader-resolver?status.svg)](https://godoc.org/github.com/Jille/raft-grpc-leader-resolver)

This small library allows you to send RPCs to your Raft leader.

It connects to all your Raft nodes and uses [client-side health checks](https://github.com/grpc/proposal/blob/master/A17-client-side-health-checking.md) to only send RPCs to the master.

During leader elections you'll see about one RTT of errors, make sure your client can handle those and retries them.

## Server side

Add this to your server:

```go
import "github.com/Jille/raft-grpc-leader-rpc/leaderhealth"

r, err := raft.NewRaft(...)
s := grpc.NewServer()

services := []string{""}
leaderhealth.Setup(r, s, services)
```

Use "" to mark all gRPC services as unhealthy if you aren't the master. Otherwise pass only the service names that you want to control healthiness for.

If you don't know what to choose, consider using the (hereby) standardized `quis.RaftLeader` service name.

Want to [read more about health checking](https://github.com/grpc/proposal/blob/master/A17-client-side-health-checking.md)?

## Client side

You need to explicitly configure your clients to look at health checks.

Add this to your client:

```go
import _ "google.golang.org/grpc/health"

c := `{"healthCheckConfig": {"serviceName": "your-service-name-or-an-empty-string"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
conn, err := grpc.Dial("dns://all-your-raft-nodes.example.com", grpc.WithDefaultServiceConfig(c))
```

Pick any of the service names you registered on the server (possibly the empty string if you used that).

You'll need to create a DNS entry that points to all your Raft nodes. If you don't feel like doing that, you can use this instead:

```go
import _ "github.com/Jille/grpc-multi-resolver"
import _ "google.golang.org/grpc/health"

c := `{"healthCheckConfig": {"serviceName": "your-service-name-or-an-empty-string"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
conn, err := grpc.Dial("multi:///127.0.0.1:50051,127.0.0.1:50052,127.0.0.1:50053", grpc.WithDefaultServiceConfig(c))
```

### Wait for Ready

I recommend enabling [Wait for Ready](https://github.com/grpc/grpc/blob/master/doc/wait-for-ready.md) by adding `grpc.WithDefaultCallOption(grpc.WithWaitForReady(true))` to your grpc.Dial(). This lets gRPC wait for a connection to the leader rather than immediately failing it if the leader is currently unknown. The deadline is still honored.
