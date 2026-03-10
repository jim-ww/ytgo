package renderer

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"

	"github.com/anthonynsimon/bild/transform"
	"github.com/mattn/go-sixel"
)

type SixelRenderer struct{}

var _ ImageRenderer = SixelRenderer{}

func (SixelRenderer) Render(r io.Reader, width, height, posX, posY int) error {
	img, err := jpeg.Decode(r)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resized := transform.Resize(img, width, height, transform.NearestNeighbor)

	var buffer bytes.Buffer

	if err := sixel.NewEncoder(&buffer).Encode(resized); err != nil {
		return fmt.Errorf("failed to encode sixel: %w", err)
	}

	fmt.Printf("\x1b7\x1b[%d;%dH%s\x1b8", posY, posX, buffer.String())
	return nil
}

func (SixelRenderer) Clear(posX, posY int) error {
	fmt.Printf("\x1b[%d;%dH\x1b[0J", posY, posX)
	return nil
}
