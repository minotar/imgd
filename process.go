package main

import (
	"github.com/disintegration/gift"
	"github.com/disintegration/imaging"
	"github.com/minotar/minecraft"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
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
func (skin *mcSkin) GetHead(width int) error {
	skin.Processed = skin.cropHead(skin.Image)
	skin.Resize(width, imaging.NearestNeighbor)
	return nil
}

// Returns the face of the skin overlayed with the helmet texture.
func (skin *mcSkin) GetHelm(width int) error {
	skin.Processed = skin.cropHelm(skin.Image)
	skin.Resize(width, imaging.NearestNeighbor)
	return nil
}

// Returns the head, torso, and arms part of the body image.
func (skin *mcSkin) RenderUpperBody() error {
	helmImg := skin.cropHead(skin.Image)
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
func (skin *mcSkin) GetBust(width int) error {
	// Go get the upper body but not all of it.
	err := skin.RenderUpperBody()
	if err != nil {
		return err
	}

	// Slice off the last little tidbit of the image.
	img := skin.Processed.(*image.NRGBA)
	img.Rect.Max.Y = BustHeight

	skin.Resize(width, imaging.NearestNeighbor)

	return nil
}

func (skin *mcSkin) GetCube(width int) error {
	// Crop out the top of the head
	topFlat := imaging.Crop(skin.Image, image.Rect(8, 0, 16, 8))
	// Resize appropriately, so that it fills the `width` when rotated 45 def.
	topFlat = imaging.Resize(topFlat, int(float64(width)*math.Sqrt(2)/3+1), 0, imaging.NearestNeighbor)
	// Create the Gift filter
	filter := gift.New(
		gift.Rotate(45, color.Transparent, gift.LinearInterpolation),
	)
	bounds := filter.Bounds(topFlat.Bounds())
	top := image.NewNRGBA(bounds)
	// Draw it on the filter, then smush it!
	filter.Draw(top, topFlat)
	top = imaging.Resize(top, width+2, width/3, imaging.NearestNeighbor)
	// Skew the front and sides at 15 degree angles to match up with the
	// head that has been smushed
	front := skin.cropHead(skin.Image).(*image.NRGBA)
	side := imaging.Crop(skin.Image, image.Rect(0, 8, 8, 16))
	front = imaging.Resize(front, width/2, int(float64(width)/1.75), imaging.NearestNeighbor)
	side = imaging.Resize(side, width/2, int(float64(width)/1.75), imaging.NearestNeighbor)
	front = skewVertical(front, math.Pi/12)
	side = skewVertical(imaging.FlipH(side), math.Pi/-12)

	// Create a new image to assemble upon
	skin.Processed = image.NewNRGBA(image.Rect(0, 0, width, width))
	// Draw each side
	draw.Draw(skin.Processed.(draw.Image), image.Rect(0, width/6, width/2, width), side, image.Pt(0, 0), draw.Src)
	draw.Draw(skin.Processed.(draw.Image), image.Rect(width/2, width/6, width, width), front, image.Pt(0, 0), draw.Src)
	// Draw the top we created
	draw.Draw(skin.Processed.(draw.Image), image.Rect(-1, 0, width+1, width/3), top, image.Pt(0, 0), draw.Over)

	return nil
}

func (skin *mcSkin) GetBody(width int) error {
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
	skin.Resize(width, imaging.NearestNeighbor)

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
func (skin *mcSkin) Resize(width int, filter imaging.ResampleFilter) {
	skin.Processed = imaging.Resize(skin.Processed, width, 0, filter)
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

func skewVertical(src *image.NRGBA, degrees float64) *image.NRGBA {
	bounds := src.Bounds()
	maxY := bounds.Max.Y
	maxX := bounds.Max.X * 4
	distance := float64(bounds.Max.X) * math.Tan(degrees)
	shouldFlip := false
	if distance < 0 {
		distance = -distance
		shouldFlip = true
	}

	newHeight := maxY + int(1+distance)
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Max.X, newHeight))

	step := distance
	for x := 0; x < maxX; x += 4 {
		for row := 0; row < maxY; row += 1 {
			srcPx := row*src.Stride + x
			dstLower := (int(step)+row)*dst.Stride + x
			dstUpper := dstLower + dst.Stride
			_, delta := math.Modf(step)

			if src.Pix[srcPx+3] != 0 {
				dst.Pix[dstLower+0] += uint8(float64(src.Pix[srcPx+0]) * (1 - delta))
				dst.Pix[dstLower+1] += uint8(float64(src.Pix[srcPx+1]) * (1 - delta))
				dst.Pix[dstLower+2] += uint8(float64(src.Pix[srcPx+2]) * (1 - delta))
				dst.Pix[dstLower+3] += uint8(float64(src.Pix[srcPx+3]) * (1 - delta))

				dst.Pix[dstUpper+0] += uint8(float64(src.Pix[srcPx+0]) * delta)
				dst.Pix[dstUpper+1] += uint8(float64(src.Pix[srcPx+1]) * delta)
				dst.Pix[dstUpper+2] += uint8(float64(src.Pix[srcPx+2]) * delta)
				dst.Pix[dstUpper+3] += uint8(float64(src.Pix[srcPx+3]) * delta)
			}
		}

		step -= distance / float64(bounds.Max.X)
	}

	if shouldFlip {
		return imaging.FlipH(dst)
	} else {
		return dst
	}
}
