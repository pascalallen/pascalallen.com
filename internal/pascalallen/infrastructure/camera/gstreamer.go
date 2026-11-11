package camera

import (
	"bytes"
	"context"
	"io"
	"log"
	"os/exec"
	"sync"
)

const gstreamerBinary = "gst-launch-1.0"

func GStreamerAvailable() bool {
	_, err := exec.LookPath(gstreamerBinary)

	return err == nil
}

func GStreamerSource(ctx context.Context) (io.ReadCloser, error) {
	path, err := exec.LookPath(gstreamerBinary)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, path,
		"-q",
		"libcamerasrc",
		"!", "video/x-raw,format=NV12,width=1280,height=720,framerate=15/1",
		"!", "videoconvert",
		"!", "jpegenc", "quality=75",
		"!", "multipartmux", "boundary=frame",
		"!", "fdsink", "fd=1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &process{cmd: cmd, stdout: stdout, stderr: stderr}, nil
}

type process struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr *bytes.Buffer
	once   sync.Once
}

func (p *process) Read(b []byte) (int, error) {
	return p.stdout.Read(b)
}

func (p *process) Close() error {
	p.once.Do(func() {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
		if out := p.stderr.String(); out != "" {
			log.Printf("camera hub: gstreamer stderr:\n%s", out)
		}
	})

	return nil
}
