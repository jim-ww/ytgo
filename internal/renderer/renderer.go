package renderer

import "io"

type ImageRenderer interface {
	Render(r io.Reader, width, height, posX, posY int) error
	Clear(posX, posY int) error
}
