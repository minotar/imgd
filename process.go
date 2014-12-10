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

func (skin *mcSkin) GetHead() error {
	skin.Processed = cropHead(skin.Image)
	return nil
}

func (skin *mcSkin) GetHelm() error {
	skin.Processed = cropHelm(skin.Image)
	return nil
}

func (skin *mcSkin) RenderUpperBody(all bool) error {
	UpperBodyHeight := HeadHeight + TorsoHeight
	UpperBodyShift := TorsoHeight
	if !all {
		//We will make a smaller UpperBody
		UpperBodyHeight = BustHeight
		UpperBodyShift = BustHeight - HeadHeight
	}

	helmImg := cropHelm(skin.Image)
	torsoImg := imaging.Crop(skin.Image, image.Rect(TorsoX, TorsoY, TorsoX+TorsoWidth, TorsoY+UpperBodyShift))
	raImg := imaging.Crop(skin.Image, image.Rect(RaX, RaY, RaX+RaWidth, RaY+UpperBodyShift))

	var laImg image.Image

	// If the skin is 1.8 then we will use the left arms and legs, otherwise flip the right ones and use them.
	if skin.is18Skin() {
		laImg = imaging.Crop(skin.Image, image.Rect(LaX, LaY, LaX+LaWidth, LaY+UpperBodyShift))
	} else {
		laImg = imaging.FlipH(raImg)
	}

	// Create a blank canvas for us to draw our upper body on
	upperBodyImg := image.NewNRGBA(image.Rect(0, 0, LaWidth+TorsoWidth+RaWidth, UpperBodyHeight))
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

func (skin *mcSkin) GetBust() error {
	// Go get the upper body but not all of it.
	err := skin.RenderUpperBody(false)
	if err != nil {
		return err
	}

	return nil
}

func (skin *mcSkin) GetBody() error {
	// Go get the upper body (all of it).
	err := skin.RenderUpperBody(true)
	if err != nil {
		return err
	}

	rlImg := imaging.Crop(skin.Image, image.Rect(RlX, RlY, RlX+RlWidth, RlY+RlHeight))

	var llImg image.Image

	// If the skin is 1.8 then we will use the left arms and legs, otherwise flip the right ones and use them.
	if skin.is18Skin() {
		llImg = imaging.Crop(skin.Image, image.Rect(LlX, LlY, LlX+LlWidth, LlY+LlHeight))
	} else {
		llImg = imaging.FlipH(rlImg)
	}

	// Create a blank canvas for us to draw our body on
	bodyImg := image.NewNRGBA(image.Rect(0, 0, LaWidth+TorsoWidth+RaWidth, HeadHeight+TorsoHeight+LlHeight))
	// Upper Body
	fastDraw(bodyImg, skin.Processed.(*image.NRGBA), 0, 0)
	// Left Leg
	fastDraw(bodyImg, llImg.(*image.NRGBA), LaWidth, HelmHeight+TorsoHeight)
	// Right Leg
	fastDraw(bodyImg, rlImg, LaWidth+LlWidth, HelmHeight+TorsoHeight)

	skin.Processed = bodyImg
	return nil
}

func (skin *mcSkin) is18Skin() bool {
	bounds := skin.Image.Bounds()
	if bounds.Max.Y == 64 {
		return true
	}
	return false
}

func (skin *mcSkin) WritePNG(w io.Writer) error {
	return png.Encode(w, skin.Processed)
}

func (skin *mcSkin) WriteSkin(w io.Writer) error {
	return png.Encode(w, skin.Image)
}

func (skin *mcSkin) Resize(width uint) {
	skin.Processed = imaging.Resize(skin.Processed, int(width), 0, imaging.NearestNeighbor)
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
