package main

import (
	"github.com/disintegration/imaging"
	"github.com/minotar/minecraft"
	"image"
	"image/png"
	"io"
)

const (
	HeadX      = 8
	HeadY      = 8
	HeadWidth  = 8
	HeadHeight = 8

	HelmX      = 40
	HelmY      = 8
	HelmWidth  = 8
	HelmHeight = 8

	TorsoX      = 20
	TorsoY      = 20
	TorsoWidth  = 8
	TorsoHeight = 12

	RaX      = 44
	RaY      = 20
	RaWidth  = 4
	RaHeight = 12

	RlX      = 4
	RlY      = 20
	RlWidth  = 4
	RlHeight = 12

	LaX      = 36
	LaY      = 52
	LaWidth  = 4
	LaHeight = 12

	LlX      = 20
	LlY      = 52
	LlWidth  = 4
	LlHeight = 12

	// The height of the 'bust' relative to the width of the body (16)
	BustHeight = 16
)

type mcSkin struct {
	Processed image.Image
	minecraft.Skin
}

// Returns the "face" of the skin.
func (skin *mcSkin) GetHead() error {
	skin.Processed = skin.cropHead(skin.Image)
	return nil
}

// Returns the face of the skin overlayed with the helmet texture.
func (skin *mcSkin) GetHelm() error {
	skin.Processed = skin.cropHelm(skin.Image)
	return nil
}

// Returns the head, torso, and arms part of the body image.
func (skin *mcSkin) RenderUpperBody() error {
	helmImg := skin.cropHelm(skin.Image)
	torsoImg := imaging.Crop(skin.Image, image.Rect(TorsoX, TorsoY, TorsoX+TorsoWidth, TorsoY+TorsoHeight))
	raImg := imaging.Crop(skin.Image, image.Rect(RaX, RaY, RaX+RaWidth, RaY+TorsoHeight))

	var laImg image.Image

	// If the skin is 1.8 then we will use the left arm, otherwise
	// flip the right ones and use them.
	if skin.is18Skin() {
		laImg = imaging.Crop(skin.Image, image.Rect(LaX, LaY, LaX+LaWidth, LaY+TorsoHeight))
	} else {
		laImg = imaging.FlipH(raImg)
	}

	// Create a blank canvas for us to draw our upper body on
	upperBodyImg := image.NewNRGBA(image.Rect(0, 0, LaWidth+TorsoWidth+RaWidth, HeadHeight+TorsoHeight))
	// Helm
	fastDraw(upperBodyImg, helmImg.(*image.NRGBA), LaWidth, 0)
	// Torso
	fastDraw(upperBodyImg, torsoImg, LaWidth, HelmHeight)
	// Left Arm
	fastDraw(upperBodyImg, laImg.(*image.NRGBA), 0, HelmHeight)
	// Right Arm
	fastDraw(upperBodyImg, raImg, LaWidth+TorsoWidth, HelmHeight)

	skin.Processed = upperBodyImg
	return nil
}

// Returns the upper portion of the body - like GetBody, but without the legs.
func (skin *mcSkin) GetBust() error {
	// Go get the upper body but not all of it.
	err := skin.RenderUpperBody()
	if err != nil {
		return err
	}

	// Slice off the last little tidbit of the image.
	img := skin.Processed.(*image.NRGBA)
	img.Rect.Max.Y = BustHeight

	return nil
}

func (skin *mcSkin) GetBody() error {
	// Go get the upper body (all of it).
	err := skin.RenderUpperBody()
	if err != nil {
		return err
	}

	rlImg := imaging.Crop(skin.Image, image.Rect(RlX, RlY, RlX+RlWidth, RlY+RlHeight))

	// If the skin is 1.8 then we will use the left arms and legs, otherwise flip the right ones and use them.
	var llImg image.Image
	if skin.is18Skin() {
		llImg = imaging.Crop(skin.Image, image.Rect(LlX, LlY, LlX+LlWidth, LlY+LlHeight))
	} else {
		llImg = imaging.FlipH(rlImg)
	}

	// Create a blank canvas for us to draw our body on. Expand bodyImg so
	// that we can draw on our legs.
	bodyImg := skin.Processed.(*image.NRGBA)
	bodyImg.Pix = append(bodyImg.Pix, make([]uint8, LlHeight*bodyImg.Stride)...)
	bodyImg.Rect.Max.Y += LlHeight
	// Left Leg
	fastDraw(bodyImg, llImg.(*image.NRGBA), LaWidth, HelmHeight+TorsoHeight)
	// Right Leg
	fastDraw(bodyImg, rlImg, LaWidth+LlWidth, HelmHeight+TorsoHeight)

	skin.Processed = bodyImg
	return nil
}

// Writes the *processed* image as a PNG to the given writer.
func (skin *mcSkin) WritePNG(w io.Writer) error {
	return png.Encode(w, skin.Processed)
}

// Writes the *original* skin image as a png to the given writer.
func (skin *mcSkin) WriteSkin(w io.Writer) error {
	return png.Encode(w, skin.Image)
}

// Resizes the skin to the given dimensions, keeping aspect ratio.
func (skin *mcSkin) Resize(width uint) {
	skin.Processed = imaging.Resize(skin.Processed, int(width), 0, imaging.NearestNeighbor)
}

// Removes the skin's alpha matte from the given image.
func (skin *mcSkin) removeAlpha(img *image.NRGBA) {
	// If it's already a transparent image, do nothing
	if skin.AlphaSig[3] == 0 {
		return
	}

	// Otherwise loop through all the pixels. Check to see which ones match
	// the alpha signature and set their opacity to be zero.
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i+0] == skin.AlphaSig[0] &&
			img.Pix[i+1] == skin.AlphaSig[1] &&
			img.Pix[i+2] == skin.AlphaSig[2] &&
			img.Pix[i+3] == skin.AlphaSig[3] {
			img.Pix[i+3] = 0
		}
	}
}

// Checks if the skin is a 1.8 skin using its height.
func (skin *mcSkin) is18Skin() bool {
	bounds := skin.Image.Bounds()
	return bounds.Max.Y == 64
}

// Returns the head of the skin image.
func (skin *mcSkin) cropHead(img image.Image) image.Image {
	return imaging.Crop(img, image.Rect(HeadX, HeadY, HeadX+HeadWidth, HeadY+HeadHeight))
}

// Returns the head of the skin image overlayed with the helm.
func (skin *mcSkin) cropHelm(img image.Image) image.Image {
	headImg := skin.cropHead(img)
	helmImg := imaging.Crop(img, image.Rect(HelmX, HelmY, HelmX+HelmWidth, HelmY+HelmHeight))
	skin.removeAlpha(helmImg)
	fastDraw(headImg.(*image.NRGBA), helmImg, 0, 0)

	return headImg
}

// Draws the "src" onto the "dst" image at the given x/y bounds, maintaining
// the original size. Pixels with have an alpha of 0x00 are not draw, and
// all others are drawn with an alpha of 0xFF
func fastDraw(dst *image.NRGBA, src *image.NRGBA, x, y int) {
	bounds := src.Bounds()
	maxY := bounds.Max.Y
	maxX := bounds.Max.X * 4

	pointer := dst.PixOffset(x, y)
	for row := 0; row < maxY; row += 1 {
		for i := 0; i < maxX; i += 4 {
			srcPx := row*src.Stride + i
			dstPx := row*dst.Stride + i + pointer
			if src.Pix[srcPx+3] != 0 {
				dst.Pix[dstPx+0] = src.Pix[srcPx+0]
				dst.Pix[dstPx+1] = src.Pix[srcPx+1]
				dst.Pix[dstPx+2] = src.Pix[srcPx+2]
				dst.Pix[dstPx+3] = 0xFF
			}
		}
	}
}
