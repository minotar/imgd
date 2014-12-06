package main

import (
	"errors"
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
)

type mcSkin struct {
	Processed image.Image
	minecraft.Skin
}

func (skin *mcSkin) GetHead() error {
	img, err := cropHead(skin.Image)
	if err != nil {
		return err
	}

	skin.Processed = img
	return nil
}

func (skin *mcSkin) GetHelm() error {
	// check if helm is solid colour - if so, it counts as transparent
	isSolidColour := true
	baseColour := skin.Image.At(HelmX, HelmY)
	for checkX := HelmX; checkX < HelmX+HelmWidth; checkX++ {
		for checkY := HelmY; checkY < HelmY+HelmHeight; checkY++ {
			checkColour := skin.Image.At(checkX, checkY)
			if checkColour != baseColour {
				isSolidColour = false
				break
			}
		}
	}

	headImg, err := cropHead(skin.Image)
	if err != nil {
		return err
	}

	skin.Processed = headImg

	if isSolidColour {
		return nil
	}

	headImgRGBA := headImg.(*image.RGBA)

	helmImg, err := cropImage(skin.Image, image.Rect(HelmX, HelmY, HelmX+HelmWidth, HelmY+HelmHeight))
	if err != nil {
		return err
	}

	sr := helmImg.Bounds()
	draw.Draw(headImgRGBA, sr, helmImg, sr.Min, draw.Over)
	return nil
}

func (skin *mcSkin) GetBody() error {
	// Check if 1.8 skin (the max Y bound should be 64)
	render18Skin := true
	bounds := skin.Image.Bounds()
	if bounds.Max.Y != 64 {
		render18Skin = false
	}

	err := skin.GetHelm()
	if err != nil {
		return err
	}
	helmImg := skin.Processed

	torsoImg, err := cropImage(skin.Image, image.Rect(TorsoX, TorsoY, TorsoX+TorsoWidth, TorsoY+TorsoHeight))
	if err != nil {
		return err
	}

	raImg, err := cropImage(skin.Image, image.Rect(RaX, RaY, RaX+RaWidth, RaY+RaHeight))
	if err != nil {
		return err
	}

	rlImg, err := cropImage(skin.Image, image.Rect(RlX, RlY, RlX+RlWidth, RlY+RlHeight))
	if err != nil {
		return err
	}

	var laImg, llImg image.Image

	// If the skin is 1.8 then we will use the left arms and legs, otherwise flip the right ones and use them.
	if render18Skin {
		laImg, err = cropImage(skin.Image, image.Rect(LaX, LaY, LaX+LaWidth, LaY+LaHeight))
		if err != nil {
			return err
		}

		llImg, err = cropImage(skin.Image, image.Rect(LlX, LlY, LlX+LlWidth, LlY+LlHeight))
		if err != nil {
			return err
		}
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

func (skin *mcSkin) WritePNG(w io.Writer) error {
	return png.Encode(w, skin.Processed)
}

func (skin *mcSkin) Resize(width uint) {
	if skin.Processed == nil {
		print("oh no")
		return
	}
	skin.Processed = imaging.Resize(skin.Processed, int(width), 0, imaging.NearestNeighbor)
}

func cropHead(img image.Image) (image.Image, error) {
	return cropImage(img, image.Rect(HeadX, HeadY, HeadX+HeadWidth, HeadY+HeadHeight))
}

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
