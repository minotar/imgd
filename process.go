package main

import (
	"github.com/disintegration/imaging"
	"github.com/minotar/minecraft"
	"image"
	"image/draw"
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

func (skin *mcSkin) GetBody() error {
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
	bodyImg := image.NewRGBA(image.Rect(0, 0, LaWidth+TorsoWidth+RaWidth, HeadHeight+TorsoHeight+LlHeight))
	// Helm
	draw.Draw(bodyImg, image.Rect(LaWidth, 0, LaWidth+HelmWidth, HelmHeight), helmImg, image.Pt(0, 0), draw.Src)
	// Torso
	draw.Draw(bodyImg, image.Rect(LaWidth, HelmHeight, LaWidth+TorsoWidth, HelmHeight+TorsoHeight), torsoImg, image.Pt(0, 0), draw.Src)
	// Left Arm
	draw.Draw(bodyImg, image.Rect(0, HelmHeight, LaWidth, HelmHeight+LaHeight), laImg, image.Pt(0, 0), draw.Src)
	// Right Arm
	draw.Draw(bodyImg, image.Rect(LaWidth+TorsoWidth, HelmHeight, LaWidth+TorsoWidth+RaWidth, HelmHeight+RaHeight), raImg, image.Pt(0, 0), draw.Src)
	// Left Leg
	draw.Draw(bodyImg, image.Rect(LaWidth, HelmHeight+TorsoHeight, LaWidth+LlWidth, HelmHeight+TorsoHeight+LlHeight), llImg, image.Pt(0, 0), draw.Src)
	// Right Leg
	draw.Draw(bodyImg, image.Rect(LaWidth+LlWidth, HelmHeight+TorsoHeight, LaWidth+LlWidth+RlWidth, HelmHeight+TorsoHeight+RlHeight), rlImg, image.Pt(0, 0), draw.Src)

	skin.Processed = bodyImg
	return nil
}

func (skin *mcSkin) GetBust() error {
	err := skin.GetBody()
	if err != nil {
		return err
	}

	skin.Processed = imaging.Crop(skin.Processed, image.Rect(0, 0, 16, 16))
	return nil
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

	sr := helmImg.Bounds()
	draw.Draw(helmImg, sr, helmImg, sr.Min, draw.Over)

	return headImg
}
