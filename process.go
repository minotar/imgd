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

func (skin *mcSkin) GetHead(width int) error {
	skin.Processed = cropHead(skin.Image)
	skin.Resize(width, imaging.NearestNeighbor)
	return nil
}

func (skin *mcSkin) GetHelm(width int) error {
	skin.Processed = cropHelm(skin.Image)
	skin.Resize(width, imaging.NearestNeighbor)
	return nil
}

func (skin *mcSkin) GetBody(width int) error {
	// Check if 1.8 skin (the max Y bound should be 64)
	render18Skin := true
	bounds := skin.Image.Bounds()
	if bounds.Max.Y != 64 {
		render18Skin = false
	}

	helmImg := cropHelm(skin.Image)
	torsoImg := imaging.Crop(skin.Image, image.Rect(TorsoX, TorsoY, TorsoX+TorsoWidth, TorsoY+TorsoHeight))
	raImg := imaging.Crop(skin.Image, image.Rect(RaX, RaY, RaX+RaWidth, RaY+RaHeight))
	rlImg := imaging.Crop(skin.Image, image.Rect(RlX, RlY, RlX+RlWidth, RlY+RlHeight))

	var laImg, llImg image.Image

	// If the skin is 1.8 then we will use the left arms and legs, otherwise flip the right ones and use them.
	if render18Skin {
		laImg = imaging.Crop(skin.Image, image.Rect(LaX, LaY, LaX+LaWidth, LaY+LaHeight))
		llImg = imaging.Crop(skin.Image, image.Rect(LlX, LlY, LlX+LlWidth, LlY+LlHeight))
	} else {
		laImg = imaging.FlipH(raImg)

		llImg = imaging.FlipH(rlImg)
	}

	// Create a blank canvas for us to draw our body on
	bodyImg := image.NewNRGBA(image.Rect(0, 0, LaWidth+TorsoWidth+RaWidth, HeadHeight+TorsoHeight+LlHeight))
	// Helm
	fastDraw(bodyImg, helmImg.(*image.NRGBA), LaWidth, 0)
	// Torso
	fastDraw(bodyImg, torsoImg, LaWidth, HelmHeight)
	// Left Arm
	fastDraw(bodyImg, laImg.(*image.NRGBA), 0, HelmHeight)
	// Right Arm
	fastDraw(bodyImg, raImg, LaWidth+TorsoWidth, HelmHeight)
	// Left Leg
	fastDraw(bodyImg, llImg.(*image.NRGBA), LaWidth, HelmHeight+TorsoHeight)
	// Right Leg
	fastDraw(bodyImg, rlImg, LaWidth+LlWidth, HelmHeight+TorsoHeight)

	skin.Processed = bodyImg
	skin.Resize(width, imaging.NearestNeighbor)
	return nil
}

func (skin *mcSkin) GetBust(width int) error {
	err := skin.GetBody(LaWidth + TorsoWidth + RaWidth)
	if err != nil {
		return err
	}

	skin.Processed = imaging.Crop(skin.Processed, image.Rect(0, 0, 16, 16))
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
	front := cropHead(skin.Image).(*image.NRGBA)
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

func (skin *mcSkin) WritePNG(w io.Writer) error {
	return png.Encode(w, skin.Processed)
}

func (skin *mcSkin) WriteSkin(w io.Writer) error {
	return png.Encode(w, skin.Image)
}

func (skin *mcSkin) Resize(width int, filter imaging.ResampleFilter) {
	skin.Processed = imaging.Resize(skin.Processed, width, 0, filter)
}

func cropHead(img image.Image) image.Image {
	return imaging.Crop(img, image.Rect(HeadX, HeadY, HeadX+HeadWidth, HeadY+HeadHeight))
}

func cropHelm(img image.Image) image.Image {
	headImg := cropHead(img)
	helmImg := imaging.Crop(img, image.Rect(HelmX, HelmY, HelmX+HelmWidth, HelmY+HelmHeight))

	fastDraw(headImg.(*image.NRGBA), helmImg, 0, 0)

	return headImg
}

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
