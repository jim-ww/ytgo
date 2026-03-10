package player

import (
	"os/exec"

	"github.com/jim-ww/ytgo/internal/types"
)

type Player interface {
	Play(v *types.Video, audioOnly bool) (*exec.Cmd, error)
	IsAvailable() bool
}
