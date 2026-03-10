package player

import (
	"os"
	"os/exec"

	"github.com/jim-ww/ytgo/internal/types"
)

type MpvPlayer struct {
	termOutput bool
}

func NewMpvPlayer(termOutput bool) MpvPlayer {
	return MpvPlayer{termOutput: termOutput}
}

var _ Player = MpvPlayer{}

func (m MpvPlayer) Play(v *types.Video, audioOnly bool) (*exec.Cmd, error) {
	args := []string{v.URL, "--really-quiet", "--msg-level=all=error"}

	if m.termOutput && !audioOnly {
		args = append(args, "-vo", "tct")
	}

	if audioOnly {
		args = append(args, "--no-video")
	}

	cmd := exec.Command("mpv", args...)

	if m.termOutput && !audioOnly {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}

	return cmd, cmd.Start()
}

func (MpvPlayer) IsAvailable() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}
