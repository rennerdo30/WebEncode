package main

import (
	"context"

	"github.com/hashicorp/go-plugin"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/ffmpeg"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type FFmpegEncoder struct {
	pb.UnimplementedEncoderServiceServer
	encoder EncoderBackend
}

func NewFFmpegEncoder() *FFmpegEncoder {
	// Simple logger to stdout for plugin
	l := logger.New("plugin-ffmpeg")
	return &FFmpegEncoder{
		encoder: &FFmpegEncoderWrapper{ffmpeg.NewFFmpegEncoder("ffmpeg", "ffprobe", l)},
	}
}

func (e *FFmpegEncoder) GetCapabilities(ctx context.Context, req *pb.Empty) (*pb.Capabilities, error) {
	// In a real implementation we would call e.encoder.GetCapabilities()
	return &pb.Capabilities{
		VideoCodecs: []string{"h264", "hevc", "vp9", "av1"},
		AudioCodecs: []string{"aac", "mp3", "opus"},
		Containers:  []string{"mp4", "mkv", "hls"},
	}, nil
}

func (e *FFmpegEncoder) Transcode(req *pb.TranscodeRequest, stream pb.EncoderService_TranscodeServer) error {
	ctx := stream.Context()

	// Map Proto request to FFmpeg task
	task := &ffmpeg.TranscodeTask{
		InputURL:    req.SourceUrl,
		OutputURL:   req.OutputUrl,
		VideoCodec:  req.Profile.VideoCodec,
		AudioCodec:  req.Profile.AudioCodec,
		BitrateKbps: int(req.Profile.Bitrate / 1000),
		Width:       int(req.Profile.Width),
		Height:      int(req.Profile.Height),
		Container:   req.Profile.Container,
		Preset:      req.Profile.Preset,
		// Duration/StartTime handling if needed
	}

	progressCh := make(chan ffmpeg.Progress)

	// Stream progress back to kernel
	go func() {
		for p := range progressCh {
			stream.Send(&pb.TranscodeProgress{
				Percent: float32(p.Percent),
				Speed:   p.Speed,
			})
		}
	}()

	if err := e.encoder.Transcode(ctx, task, progressCh); err != nil {
		return err
	}

	// Send final completion if not sent
	stream.Send(&pb.TranscodeProgress{Percent: 100, Completed: true})
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{
				EncoderImpl: NewFFmpegEncoder(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
