package main

import (
	"github.com/gographics/imagick/imagick"
	"github.com/minotar/minecraft"
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
	Processed *imagick.MagickWand
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
	if skin.Image.GetImageHeight() != 64 {
		render18Skin = false
	}

	helmImg := cropHelm(skin.Image)
	// Create necessary subimages
	torsoImg := skin.Image.Clone()
	raImg := skin.Image.Clone()
	rlImg := skin.Image.Clone()
	var laImg, llImg *imagick.MagickWand

	// Destroy them all after we're done
	defer torsoImg.Destroy()
	defer raImg.Destroy()
	defer rlImg.Destroy()

	// And start croppin!
	torsoImg.CropImage(TorsoWidth, TorsoHeight, TorsoX, TorsoY)
	raImg.CropImage(RaWidth, RaHeight, RaX, RaY)
	rlImg.CropImage(RlWidth, RlHeight, RlX, RlY)

	// If the skin is 1.8 then we will use the left arms and legs, otherwise flip the right ones and use them.
	if render18Skin {
		laImg = skin.Image.Clone()
		llImg = skin.Image.Clone()
		laImg.CropImage(LaWidth, LaHeight, LaX, LaY)
		llImg.CropImage(LlWidth, LlHeight, LlX, LlY)
	} else {
		laImg = raImg.Clone()
		llImg = rlImg.Clone()
		laImg.FlopImage()
		llImg.FlopImage()
	}

	defer laImg.Destroy()
	defer llImg.Destroy()

	// Create a blank canvas for us to draw our body on
	bodyImg := imagick.NewMagickWand()
	bodyImg.SetFormat("PNG")
	bg := imagick.NewPixelWand()
	bg.SetColor("none")

	bodyImg.NewImage(LaWidth+TorsoWidth+RaWidth, HeadHeight+TorsoHeight+LlHeight, bg)
	// Helm
	bodyImg.CompositeImage(helmImg, imagick.COMPOSITE_OP_OVER, LaWidth, 0)
	// Torso
	bodyImg.CompositeImage(torsoImg, imagick.COMPOSITE_OP_OVER, LaWidth, HelmHeight)
	// Left Arm
	bodyImg.CompositeImage(laImg, imagick.COMPOSITE_OP_OVER, 0, HelmHeight)
	// Right Arm
	bodyImg.CompositeImage(raImg, imagick.COMPOSITE_OP_OVER, LaWidth+TorsoWidth, HelmHeight)
	// Left Leg
	bodyImg.CompositeImage(llImg, imagick.COMPOSITE_OP_OVER, LaWidth, HelmHeight+TorsoHeight)
	// Right Leg
	bodyImg.CompositeImage(rlImg, imagick.COMPOSITE_OP_OVER, LaWidth+LlWidth, HelmHeight+TorsoHeight)

	skin.Processed = bodyImg
	return nil
}

func (skin *mcSkin) GetBust() error {
	err := skin.GetBody()
	if err != nil {
		return err
	}

	body := skin.Processed
	skin.Processed = body.Clone()
	defer body.Destroy()

	body.CropImage(0, 0, 16, 16)

	return nil
}

func (skin *mcSkin) WritePNG(w io.Writer) {
	w.Write(skin.Processed.GetImageBlob())
}

func (skin *mcSkin) WriteSkin(w io.Writer) {
	w.Write(skin.Processed.GetImageBlob())
}

func (skin *mcSkin) Destroy() {
	if skin.Processed != nil {
		skin.Processed.Destroy()
	}
	if skin.Image != nil {
		skin.Image.Destroy()
	}
}

func (skin *mcSkin) Resize(targetWidth uint) {
	width := skin.Processed.GetImageWidth()
	height := skin.Processed.GetImageHeight()
	scaled := float64(targetWidth) / float64(width)

	skin.Processed.ResizeImage(targetWidth, uint(scaled*float64(height)), imagick.FILTER_BOX, 0)
}

func cropHead(img *imagick.MagickWand) *imagick.MagickWand {
	out := img.Clone()
	out.CropImage(HeadWidth, HeadHeight, HeadX, HeadY)
	return out
}

func cropHelm(img *imagick.MagickWand) *imagick.MagickWand {
	helmImg := img.Clone()
	helmImg.CropImage(HelmWidth, HelmHeight, HelmX, HelmY)
	defer helmImg.Destroy()

	headImg := cropHead(img)
	headImg.CompositeImage(helmImg, imagick.COMPOSITE_OP_OVER, 0, 0)

	return headImg
}
