/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"d7y.io/dragonfly/v2/pkg/rpc"
	"d7y.io/dragonfly/v2/pkg/rpc/scheduler"
	"d7y.io/dragonfly/v2/scheduler/metrics"
)

// SchedulerServer refer to scheduler.SchedulerServer
type SchedulerServer interface {
	// RegisterPeerTask registers a peer into one task.
	RegisterPeerTask(context.Context, *scheduler.PeerTaskRequest) (*scheduler.RegisterResult, error)
	// ReportPieceResult reports piece results and receives peer packets.
	ReportPieceResult(context.Context, scheduler.Scheduler_ReportPieceResultServer) error
	// ReportPeerResult reports downloading result for the peer task.
	ReportPeerResult(context.Context, *scheduler.PeerResult) error
	// LeaveTask makes the peer leaving from scheduling overlay for the task.
	LeaveTask(context.Context, *scheduler.PeerTarget) error
}

type proxy struct {
	server SchedulerServer
	scheduler.UnimplementedSchedulerServer
}

func New(schedulerServer SchedulerServer, opts ...grpc.ServerOption) *grpc.Server {
	grpcServer := grpc.NewServer(append(rpc.DefaultServerOptions, opts...)...)
	scheduler.RegisterSchedulerServer(grpcServer, &proxy{server: schedulerServer})
	return grpcServer
}

func (p *proxy) RegisterPeerTask(ctx context.Context, req *scheduler.PeerTaskRequest) (*scheduler.RegisterResult, error) {
	metrics.RegisterPeerTaskCount.Inc()
	resp, err := p.server.RegisterPeerTask(ctx, req)
	if err != nil {
		metrics.RegisterPeerTaskFailureCount.Inc()
	} else {
		metrics.PeerTaskCounter.WithLabelValues(resp.SizeScope.String()).Inc()
	}

	return resp, err
}

func (p *proxy) ReportPieceResult(stream scheduler.Scheduler_ReportPieceResultServer) error {
	metrics.ConcurrentScheduleGauge.Inc()
	defer metrics.ConcurrentScheduleGauge.Dec()
	ctx := stream.Context()
	peerAddr := "unknown"
	if pe, ok := peer.FromContext(ctx); ok {
		peerAddr = pe.Addr.String()
	}
	logger.Infof("start report piece from: %s", peerAddr)
	return p.server.ReportPieceResult(ctx, stream)
}

func (p *proxy) ReportPeerResult(ctx context.Context, req *scheduler.PeerResult) (*empty.Empty, error) {
	metrics.DownloadCount.Inc()
	if req.Success {
		metrics.P2PTraffic.Add(float64(req.Traffic))
		metrics.PeerTaskDownloadDuration.Observe(float64(req.Cost))
	} else {
		metrics.DownloadFailureCount.Inc()
	}

	return new(empty.Empty), p.server.ReportPeerResult(ctx, req)
}

func (p *proxy) LeaveTask(ctx context.Context, pt *scheduler.PeerTarget) (*empty.Empty, error) {
	return new(empty.Empty), p.server.LeaveTask(ctx, pt)
}
