package service

import "github.com/Wei-Shaw/sub2api/internal/config"

const (
	defaultUsageDetailCaptureMaxRequestBytes  int64 = 0
	defaultUsageDetailCaptureMaxResponseBytes int64 = 0
	defaultUsageDetailCaptureMaxWSFrames            = 0
)

type UsageDetailCaptureConfig struct {
	Enabled             bool
	MaxRequestBytes     int64
	MaxResponseBytes    int64
	MaxWSResponseFrames int
}

type UsageCapturedPayload struct {
	Body     []byte
	Bytes    int64
	Complete bool
}

type UsageCapturedFrames struct {
	Frames   [][]byte
	Bytes    int64
	Complete bool
}

type UsageCapturedFrameRecorder struct {
	maxBytes   int64
	maxFrames  int
	totalBytes int64
	captured   int64
	complete   bool
	frames     [][]byte
}

func ResolveUsageDetailCaptureConfig(cfg *config.Config) UsageDetailCaptureConfig {
	out := UsageDetailCaptureConfig{
		Enabled:             true,
		MaxRequestBytes:     defaultUsageDetailCaptureMaxRequestBytes,
		MaxResponseBytes:    defaultUsageDetailCaptureMaxResponseBytes,
		MaxWSResponseFrames: defaultUsageDetailCaptureMaxWSFrames,
	}
	if cfg == nil {
		return out
	}
	out.Enabled = cfg.Gateway.UsageDetailCapture.Enabled
	if cfg.Gateway.UsageDetailCapture.MaxRequestBytes >= 0 {
		out.MaxRequestBytes = cfg.Gateway.UsageDetailCapture.MaxRequestBytes
	}
	if cfg.Gateway.UsageDetailCapture.MaxResponseBytes >= 0 {
		out.MaxResponseBytes = cfg.Gateway.UsageDetailCapture.MaxResponseBytes
	}
	if cfg.Gateway.UsageDetailCapture.MaxWSResponseFrames >= 0 {
		out.MaxWSResponseFrames = cfg.Gateway.UsageDetailCapture.MaxWSResponseFrames
	}
	if !out.Enabled {
		out.MaxRequestBytes = 0
		out.MaxResponseBytes = 0
		out.MaxWSResponseFrames = 0
	}
	return out
}

func CaptureUsagePayload(raw []byte, maxBytes int64) UsageCapturedPayload {
	out := UsageCapturedPayload{
		Bytes:    int64(len(raw)),
		Complete: true,
	}
	if len(raw) == 0 {
		return out
	}
	if maxBytes == 0 || int64(len(raw)) <= maxBytes {
		out.Body = append([]byte(nil), raw...)
		return out
	}
	out.Body = append([]byte(nil), raw[:maxBytes]...)
	out.Complete = false
	return out
}

func NewUsageCapturedFrameRecorder(maxBytes int64, maxFrames int) *UsageCapturedFrameRecorder {
	return &UsageCapturedFrameRecorder{
		maxBytes:  maxBytes,
		maxFrames: maxFrames,
		complete:  true,
	}
}

func (r *UsageCapturedFrameRecorder) Add(payload []byte) {
	if r == nil {
		return
	}

	r.totalBytes += int64(len(payload))
	if len(payload) == 0 {
		return
	}
	if r.maxFrames > 0 && len(r.frames) >= r.maxFrames {
		r.complete = false
		return
	}

	captureLen := int64(len(payload))
	if r.maxBytes > 0 {
		remaining := r.maxBytes - r.captured
		if remaining <= 0 {
			r.complete = false
			return
		}
		if captureLen > remaining {
			captureLen = remaining
		}
	}
	frame := append([]byte(nil), payload[:captureLen]...)
	r.frames = append(r.frames, frame)
	r.captured += captureLen
	if captureLen < int64(len(payload)) || (r.maxBytes > 0 && r.captured >= r.maxBytes) {
		r.complete = false
	}
}

func (r *UsageCapturedFrameRecorder) Snapshot() UsageCapturedFrames {
	if r == nil {
		return UsageCapturedFrames{Complete: true}
	}
	frames := make([][]byte, 0, len(r.frames))
	for idx := range r.frames {
		frames = append(frames, append([]byte(nil), r.frames[idx]...))
	}
	return UsageCapturedFrames{
		Frames:   frames,
		Bytes:    r.totalBytes,
		Complete: r.complete,
	}
}
