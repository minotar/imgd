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
)

func GetHead(skin minecraft.Skin) (image.Image, error) {
	return cropImage(skin.Image, image.Rect(HeadX, HeadY, HeadX+HeadWidth, HeadY+HeadHeight))
}

func GetHelm(skin minecraft.Skin) (image.Image, error) {
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

	if isSolidColour {
		return GetHead(skin)
	}

	headImg, err := GetHead(skin)
	if err != nil {
		return nil, err
	}

	headImgRGBA := headImg.(*image.RGBA)

	helmImg, err := cropImage(skin.Image, image.Rect(HelmX, HelmY, HelmX+HelmWidth, HelmY+HelmHeight))
	if err != nil {
		return nil, err
	}

	sr := helmImg.Bounds()
	draw.Draw(headImgRGBA, sr, helmImg, sr.Min, draw.Over)

	return headImg, nil
}

func WritePNG(w io.Writer, i image.Image) error {
	return png.Encode(w, i)
}

func Resize(width, height uint, img image.Image) image.Image {
	return imaging.Resize(img, int(width), int(height), imaging.NearestNeighbor)
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
