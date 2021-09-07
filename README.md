# raft-grpc-leader-rpc

Send gRPCs to your [Raft](https://github.com/hashicorp/raft) leader.

It connects to all your Raft nodes and uses [client-side health checks](https://github.com/grpc/proposal/blob/master/A17-client-side-health-checking.md) to only send RPCs to the master.

During leader elections there will temporarily be no leader and you'll see errors, make sure your client can handle those and retries them.

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

You need to explicitly configure your clients to enable health checking.

Add this to your client:

```go
import _ "google.golang.org/grpc/health"

c := `{"healthCheckConfig": {"serviceName": "quis.RaftLeader"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
target := "dns://all-your-raft-nodes.example.com"
conn, err := grpc.Dial(target, grpc.WithDefaultServiceConfig(c))
```

Instead of `quis.RaftLeader` you can also pick any of the service names you've registered with leaderhealth.Setup().

You can also configure the _ServiceConfig_ in DNS if you want to.

You'll need to create a DNS entry that points to all your Raft nodes.

### No DNS entry?

If you don't have a DNS entry, check out https://github.com/Jille/grpc-multi-resolver. Usage is easy.

```go
import _ "github.com/Jille/grpc-multi-resolver"

target := "multi:///127.0.0.1:50051,127.0.0.1:50052,127.0.0.1:50053"
```

### Wait for Ready

I recommend enabling [Wait for Ready](https://github.com/grpc/grpc/blob/master/doc/wait-for-ready.md) by adding `grpc.WithDefaultCallOptions(grpc.WaitForReady(true))` to your grpc.Dial(). This lets gRPC wait for a connection to the leader rather than immediately failing RPCs if the leader is currently unknown. The deadline is still honored.

If you get errors like `connection active but health check failed.`, this is what you want to enable.

## Automatic retries

You can use https://godoc.org/github.com/grpc-ecosystem/go-grpc-middleware/retry to transparently retry failures without the client code knowing it.

You should enable _Wait for Ready_, otherwise it might burn through all the retries before there is a new leader.

Add this to your client:

```go
import "github.com/grpc-ecosystem/go-grpc-middleware/retry"

retryOpts := []grpc_retry.CallOption{
	grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	grpc_retry.WithMax(5), // Give up after 5 retries.
}
grpc.Dial(..., grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
```

**Your server will need some modifications too.** Each of your RPCs needs to return an appropriate status code.

[![Godoc](https://godoc.org/github.com/Jille/raft-grpc-leader-rpc/rafterrors?status.svg)](https://godoc.org/github.com/Jille/raft-grpc-leader-rpc/rafterrors)

Make sure to read `rafterrors`' documentation to know when to use MarkRetriable vs MarkUnretriable, there's a pitfall.
