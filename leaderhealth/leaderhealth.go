// Package leaderhealth is included in your Raft nodes to expose whether this node is the leader.
package leaderhealth

import (
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Setup creates a new health.Server for you and registers it on s.
// It's a convenience wrapper around Report.
func Setup(r *raft.Raft, s *grpc.Server, services []string) {
	hs := health.NewServer()
	Report(r, hs, services)
	grpc_health_v1.RegisterHealthServer(s, hs)
}

// Report starts a goroutine that updates the given health.Server with whether we are the Raft leader.
// It will set the given services as SERVING if we are the leader, and as NOT_SERVING otherwise.
func Report(r *raft.Raft, hs *health.Server, services []string) {
	lch1 := r.LeaderCh()
	lch2 := r.LeaderCh()
	if lch1 == lch2 {
		ch := make(chan raft.Observation, 1)
		r.RegisterObserver(raft.NewObserver(ch, true, func(o *raft.Observation) bool {
			_, ok := o.Data.(raft.LeaderObservation)
			return ok
		}))
		setServingStatus(hs, services, r.State() == raft.Leader)
		go func() {
			for range ch {
				setServingStatus(hs, services, r.State() == raft.Leader)
			}
		}()
	} else {
		setServingStatus(hs, services, <-lch1)
		go func() {
			for isLeader := range lch1 {
				setServingStatus(hs, services, isLeader)
			}
		}()
	}
}

func setServingStatus(hs *health.Server, services []string, isLeader bool) {
	v := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if isLeader {
		v = grpc_health_v1.HealthCheckResponse_SERVING
	}
	for _, srv := range services {
		hs.SetServingStatus(srv, v)
	}
	hs.SetServingStatus("quis.RaftLeader", v)
}
