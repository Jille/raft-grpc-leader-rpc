# raft-grpc-leader-rpc

Send gRPCs to your [Raft](https://github.com/hashicorp/raft) leader.

It connects to all your Raft nodes and uses [client-side health checks](https://github.com/grpc/proposal/blob/master/A17-client-side-health-checking.md) to only send RPCs to the master.

During leader elections you'll see errors, make sure your client can handle those and retries them.

## Server side

[![Godoc](https://godoc.org/github.com/Jille/raft-grpc-leader-rpc/leaderhealth?status.svg)](https://godoc.org/github.com/Jille/raft-grpc-leader-rpc/leaderhealth)

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

c := `{"healthCheckConfig": {"serviceName": "quis.RaftLeader"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
conn, err := grpc.Dial("dns://all-your-raft-nodes.example.com", grpc.WithDefaultServiceConfig(c))
```

Instead of `quis.RaftLeader` you can also pick any of the service names you registered with leaderhealth.Setup().

You'll need to create a DNS entry that points to all your Raft nodes.

### No DNS entry?

If you don't feel like doing that, you can use this instead:

```go
import _ "github.com/Jille/grpc-multi-resolver"
import _ "google.golang.org/grpc/health"

c := `{"healthCheckConfig": {"serviceName": "your-service-name-or-an-empty-string"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
conn, err := grpc.Dial("multi:///127.0.0.1:50051,127.0.0.1:50052,127.0.0.1:50053", grpc.WithDefaultServiceConfig(c))
```

### Wait for Ready

I recommend enabling [Wait for Ready](https://github.com/grpc/grpc/blob/master/doc/wait-for-ready.md) by adding `grpc.WithDefaultCallOptions(grpc.WaitForReady(true))` to your grpc.Dial(). This lets gRPC wait for a connection to the leader rather than immediately failing it if the leader is currently unknown. The deadline is still honored.

When you get errors like `connection active but health check failed.`, this is what you want to enable.

## Automatic retries

You can use https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/retry to transparently retry failures without the client code knowing it.

You're gonna want to enable Wait for Ready or this isn't going to make it very transparent for your clients.

Add this to your client:

```go
import "github.com/grpc-ecosystem/go-grpc-middleware/retry"

retryOpts := []grpc_retry.CallOption{
	grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	grpc_retry.WithMax(5), // Give up after 5 retries.
}
grpc.Dial(..., grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
```

Your server will need to more modifications. Each of your RPCs needs to return appropriate status codes.

[![Godoc](https://godoc.org/github.com/Jille/raft-grpc-leader-rpc/rafterrors?status.svg)](https://godoc.org/github.com/Jille/raft-grpc-leader-rpc/rafterrors)

Make sure to read rafterror's documentation to known when to use MarkRetriable vs MarkUnretriable, there's a pitfall.
