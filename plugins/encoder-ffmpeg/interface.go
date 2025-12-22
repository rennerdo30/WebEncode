package main

import (
	"context"

	"github.com/rennerdo30/webencode/pkg/ffmpeg"
)

// Define Service Interface for mocking
type EncoderBackend interface {
	Transcode(ctx context.Context, task *ffmpeg.TranscodeTask, progressCh chan<- ffmpeg.Progress) error
}

// Wrapper for existing concrete implementation
type FFmpegEncoderWrapper struct {
	*ffmpeg.FFmpegEncoder
}

func (w *FFmpegEncoderWrapper) Transcode(ctx context.Context, task *ffmpeg.TranscodeTask, progressCh chan<- ffmpeg.Progress) error {
	return w.FFmpegEncoder.Transcode(ctx, task, progressCh)
}

// Update FFmpegEncoder struct to use interface
/*
type FFmpegEncoder struct {
	pb.UnimplementedEncoderServiceServer
	encoder EncoderBackend
}
*/
