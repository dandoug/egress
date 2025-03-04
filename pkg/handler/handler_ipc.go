// Copyright 2023 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/livekit/egress/pkg/errors"
	"github.com/livekit/egress/pkg/ipc"
	"github.com/livekit/protocol/pprof"
	"github.com/livekit/protocol/tracer"
)

func (h *Handler) GetPipelineDot(ctx context.Context, _ *ipc.GstPipelineDebugDotRequest) (*ipc.GstPipelineDebugDotResponse, error) {
	ctx, span := tracer.Start(ctx, "Handler.GetPipelineDot")
	defer span.End()

	if h.pipeline == nil {
		return nil, errors.ErrEgressNotFound
	}

	res := make(chan string, 1)
	go func() {
		res <- h.pipeline.GetGstPipelineDebugDot()
	}()

	select {
	case r := <-res:
		return &ipc.GstPipelineDebugDotResponse{
			DotFile: r,
		}, nil

	case <-time.After(2 * time.Second):
		return nil, status.New(codes.DeadlineExceeded, "timed out requesting pipeline debug info").Err()
	}
}

func (h *Handler) GetPProf(ctx context.Context, req *ipc.PProfRequest) (*ipc.PProfResponse, error) {
	ctx, span := tracer.Start(ctx, "Handler.GetPProf")
	defer span.End()

	if h.pipeline == nil {
		return nil, errors.ErrEgressNotFound
	}

	b, err := pprof.GetProfileData(ctx, req.ProfileName, int(req.Timeout), int(req.Debug))
	if err != nil {
		return nil, err
	}

	return &ipc.PProfResponse{
		PprofFile: b,
	}, nil
}

// GetMetrics implement the handler-side gathering of metrics to return over IPC
func (h *Handler) GetMetrics(ctx context.Context, req *ipc.MetricsRequest) (*ipc.MetricsResponse, error) {
	ctx, span := tracer.Start(ctx, "Handler.GetMetrics")
	defer span.End()

	metricsAsString, err := h.GenerateMetrics(ctx)
	if err != nil {
		return nil, err
	}

	return &ipc.MetricsResponse{
		Metrics: metricsAsString,
	}, nil
}
