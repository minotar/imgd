package minotar

import (
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/png"
	//"image/color"
	"errors"
	"io"
)

const (
	HEAD_X      = 8
	HEAD_Y      = 8
	HEAD_WIDTH  = 8
	HEAD_HEIGHT = 8

	HELM_X      = 40
	HELM_Y      = 8
	HELM_WIDTH  = 8
	HELM_HEIGHT = 8
)

func cropImage(i image.Image, d image.Rectangle) (image.Image, error) {
	bounds := i.Bounds()
	if bounds.Min.X > d.Min.X || bounds.Min.Y > d.Min.Y || bounds.Max.X < d.Max.X || bounds.Max.Y < d.Max.Y {
		return nil, errors.New("Bounds invalid for crop")
	}

	dims := d.Size()
	outIm := image.NewRGBA(image.Rect(0, 0, dims.X, dims.Y))
	for x := 0; x < dims.X; x++ {
		for y := 0; y < dims.Y; y++ {
			outIm.Set(x, y, i.At(d.Min.X+x, d.Min.Y+y))
		}
	}
	return outIm, nil
}

type Skin struct {
	Image image.Image
}

func (i Skin) Head() (image.Image, error) {
	return cropImage(i.Image, image.Rect(HEAD_X, HEAD_Y, HEAD_X+HEAD_WIDTH, HEAD_Y+HEAD_HEIGHT))
}

func (i Skin) Helm() (image.Image, error) {
	// check if helm is solid colour - if so, it counts as transparent
	isSolidColour := true
	baseColour := i.Image.At(HELM_X, HELM_Y)
	for checkX := HELM_X; checkX < HELM_X+HELM_WIDTH; checkX++ {
		for checkY := HELM_Y; checkY < HELM_Y+HELM_HEIGHT; checkY++ {
			checkColour := i.Image.At(checkX, checkY)
			if checkColour != baseColour {
				isSolidColour = false
				break
			}
		}
	}

	if isSolidColour {
		return i.Head()
	}

	headImg, err := i.Head()
	if err != nil {
		return nil, err
	}

	headImgRGBA := headImg.(*image.RGBA)

	helmImg, err := cropImage(i.Image, image.Rect(HELM_X, HELM_Y, HELM_X+HELM_WIDTH, HELM_Y+HELM_HEIGHT))
	if err != nil {
		return nil, err
	}

	sr := helmImg.Bounds()
	draw.Draw(headImgRGBA, sr, helmImg, sr.Min, draw.Over)

	return headImg, nil
}

func DecodeSkin(r io.Reader) (Skin, error) {
	skinImg, _, err := image.Decode(r)
	if err != nil {
		return Skin{}, err
	}
	return Skin{
		Image: skinImg,
	}, err
}

func WritePNG(w io.Writer, i image.Image) error {
	return png.Encode(w, i)
}

func Resize(width, height uint, img image.Image) image.Image {
	return resize.Resize(width, height, img, resize.NearestNeighbor)
}
