// mcskin has the logic for constructing the skin into the desired shape
// It also will detect width/image type and can be used as an http.Handler to
// then deliver this to a client
package mcskin

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"net/http"
	"strconv"

	svg "github.com/ajstarks/svgo"
	"github.com/disintegration/gift"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/minecraft"
)

const (
	HeadX      = 8
	HeadY      = 8
	HeadWidth  = 8
	HeadHeight = 8

	HelmX = 40
	HelmY = 8

	TorsoX      = 20
	TorsoY      = 20
	TorsoWidth  = 8
	TorsoHeight = 12

	Torso2X = 20
	Torso2Y = 36

	RaX      = 44
	RaY      = 20
	RaWidth  = 4
	RaHeight = 12

	Ra2X = 44
	Ra2Y = 36

	RlX      = 4
	RlY      = 20
	RlWidth  = 4
	RlHeight = 12

	Rl2X = 4
	Rl2Y = 36

	LaX      = 36
	LaY      = 52
	LaWidth  = 4
	LaHeight = 12

	La2X = 52
	La2Y = 52

	LlX      = 20
	LlY      = 52
	LlWidth  = 4
	LlHeight = 12

	Ll2X = 4
	Ll2Y = 52

	// The height of the 'bust' relative to the width of the body (16)
	BustHeight = 16
)

// Set the default, min and max width to resize processed images to
const (
	DefaultWidth = int(180)
	MinWidth     = int(8)
	MaxWidth     = int(300)
)

const (
	ImageTypePNG ImageType = "image/png"
	ImageTypeSVG ImageType = "image/svg+xml"
)

type ImageType string

// GetWidth converts and sanitizes the string for the avatar width.
func GetWidth(inp string) int {
	out, err := strconv.Atoi(inp)
	if err != nil {
		return DefaultWidth
	} else if out > MaxWidth {
		return MaxWidth
	} else if out < MinWidth {
		return MinWidth
	}
	return out
}

// GetImageType checks for specific extensions, otherwise uses PNG
func GetImageType(ext string) ImageType {
	switch ext {
	case ".svg":
		return ImageTypeSVG
	default:
		return ImageTypePNG
	}
}

// Based on *http.Request, read the Gorilla Mux vars for "width" and "extension"
func GetWidthType(r *http.Request) (width int, imageType ImageType) {
	vars := mux.Vars(r)
	if reqWidth, widthGiven := vars["width"]; widthGiven {
		width = GetWidth(reqWidth)
	} else {
		width = DefaultWidth
	}

	if reqExtension, extensionGiven := vars["extension"]; extensionGiven {
		imageType = GetImageType(reqExtension)
	} else {
		imageType = ImageTypePNG
	}
	return
}

// Create a McSkin for manual usage (vs. using the Handlers)
func NewMcSkinFromRequest(r *http.Request, skin minecraft.Skin) *McSkin {
	mcSkin := &McSkin{Skin: skin}
	mcSkin.Width, mcSkin.Type = GetWidthType(r)
	return mcSkin
}

type McSkin struct {
	minecraft.Skin
	Processed image.Image
	Processor func() error
	Type      ImageType
	Width     int
}

// Sets skin.Processed to the face of the user.
func (skin *McSkin) GetHead() error {
	skin.Processed = skin.cropHead(skin.Image)
	skin.resize(imaging.NearestNeighbor)
	return nil
}

// Sets skin.Processed to the face of the user overlaid with their helmet.
func (skin *McSkin) GetHelm() error {
	skin.Processed = skin.cropHelm(skin.Image)
	skin.resize(imaging.NearestNeighbor)
	return nil
}

// Sets skin.Processed to an isometric render of the head from a top-left angle (showing 3 sides).
func (skin *McSkin) GetCube() error {
	width := skin.Width
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

// Sets skin.Processed to an isometric render of the head from a top-left angle (showing 3 sides).
func (skin *McSkin) GetCubeHelm() error {
	width := skin.Width
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

	// Crop out the top of the head
	topFlatHelm := imaging.Crop(skin.Image, image.Rect(40, 0, 48, 8))
	// Resize appropriately, so that it fills the `width` when rotated 45 def.
	topFlatHelm = imaging.Resize(topFlatHelm, int(float64(width)*math.Sqrt(2)/3+1), 0, imaging.NearestNeighbor)

	boundshelm := filter.Bounds(topFlatHelm.Bounds())
	tophelm := image.NewNRGBA(boundshelm)
	// Draw it on the filter, then smush it!
	filter.Draw(tophelm, topFlatHelm)
	tophelm = imaging.Resize(tophelm, width+2, width/3, imaging.NearestNeighbor)

	// Skew the front and sides at 15 degree angles to match up with the
	// head that has been smushed
	side := imaging.Crop(skin.Image, image.Rect(0, 8, 8, 16))
	side = imaging.Resize(side, width/2, int(float64(width)/1.75), imaging.NearestNeighbor)
	side = skewVertical(imaging.FlipH(side), math.Pi/-12)

	sidehelm := imaging.Crop(skin.Image, image.Rect(32, 8, 40, 16))
	sidehelm = imaging.Resize(sidehelm, width/2, int(float64(width)/1.75), imaging.NearestNeighbor)
	sidehelm = skewVertical(imaging.FlipH(sidehelm), math.Pi/-12)

	front := skin.cropHead(skin.Image).(*image.NRGBA)
	front = imaging.Resize(front, width/2, int(float64(width)/1.75), imaging.NearestNeighbor)
	front = skewVertical(front, math.Pi/12)

	fronthelm := skin.cropHelm(skin.Image).(*image.NRGBA)
	fronthelm = imaging.Resize(fronthelm, width/2, int(float64(width)/1.75), imaging.NearestNeighbor)
	fronthelm = skewVertical(fronthelm, math.Pi/12)

	// Create a new image to assemble upon
	skin.Processed = image.NewNRGBA(image.Rect(0, 0, width, width))
	// Draw each side
	draw.Draw(skin.Processed.(draw.Image), image.Rect(0, width/6, width/2, width), side, image.Pt(0, 0), draw.Src)
	draw.Draw(skin.Processed.(draw.Image), image.Rect(0, width/6, width/2, width), sidehelm, image.Pt(0, 0), draw.Over)

	draw.Draw(skin.Processed.(draw.Image), image.Rect(width/2, width/6, width, width), front, image.Pt(0, 0), draw.Src)
	draw.Draw(skin.Processed.(draw.Image), image.Rect(width/2, width/6, width, width), fronthelm, image.Pt(0, 0), draw.Over)
	// Draw the top we created
	draw.Draw(skin.Processed.(draw.Image), image.Rect(-1, 0, width+1, width/3), top, image.Pt(0, 0), draw.Over)
	draw.Draw(skin.Processed.(draw.Image), image.Rect(-1, 0, width+1, width/3), tophelm, image.Pt(0, 0), draw.Over)

	return nil
}

// Sets skin.Processed to the upper portion of the body (slightly higher cutoff than waist).
func (skin *McSkin) GetBust() error {
	headImg := skin.cropHead(skin.Image).(*image.NRGBA)
	upperBodyImg := skin.renderUpperBody()

	bustImg := skin.addHead(upperBodyImg, headImg)

	bustImg.Rect.Max.Y = BustHeight
	skin.Processed = bustImg

	skin.resize(imaging.NearestNeighbor)

	return nil
}

// Sets skin.Processed to the upper portion of the body (slightly higher cutoff than waist) but with any armor which the user has.
func (skin *McSkin) GetArmorBust() error {
	helmImg := skin.cropHelm(skin.Image).(*image.NRGBA)
	upperArmorImg := skin.renderUpperArmor()

	bustImg := skin.addHead(upperArmorImg, helmImg)

	bustImg.Rect.Max.Y = BustHeight
	skin.Processed = bustImg

	skin.resize(imaging.NearestNeighbor)

	return nil
}

// Sets skin.Processed to a front render of the body.
func (skin *McSkin) GetBody() error {
	headImg := skin.cropHead(skin.Image).(*image.NRGBA)
	upperBodyImg := skin.renderUpperBody()
	lowerBodyImg := skin.renderLowerBody()

	bodyImg := skin.addHead(upperBodyImg, headImg)
	skin.Processed = skin.addLegs(bodyImg, lowerBodyImg)

	skin.resize(imaging.NearestNeighbor)

	return nil
}

// Sets skin.Processed to a front render of the body but with any armor which the user has.
func (skin *McSkin) GetArmorBody() error {
	helmImg := skin.cropHelm(skin.Image).(*image.NRGBA)
	upperArmorImg := skin.renderUpperArmor()
	lowerArmorImg := skin.renderLowerArmor()

	bodyImg := skin.addHead(upperArmorImg, helmImg)
	skin.Processed = skin.addLegs(bodyImg, lowerArmorImg)

	skin.resize(imaging.NearestNeighbor)

	return nil
}

// Returns the torso and arms.
func (skin *McSkin) renderUpperBody() *image.NRGBA {
	// This will be the base.
	upperBodyImg := image.NewNRGBA(image.Rect(0, 0, LaWidth+TorsoWidth+RaWidth, TorsoHeight))

	torsoImg := imaging.Crop(skin.Image, image.Rect(TorsoX, TorsoY, TorsoX+TorsoWidth, TorsoY+TorsoHeight))
	raImg := imaging.Crop(skin.Image, image.Rect(RaX, RaY, RaX+RaWidth, RaY+TorsoHeight))

	// If it's an old skin, they don't have a Left Arm, so we'll just flip their right.
	var laImg image.Image
	if skin.is18Skin() {
		laImg = imaging.Crop(skin.Image, image.Rect(LaX, LaY, LaX+LaWidth, LaY+TorsoHeight))
	} else {
		laImg = imaging.FlipH(raImg)
	}

	return skin.drawUpper(upperBodyImg, torsoImg, raImg, laImg.(*image.NRGBA))
}

// Returns the torso and arms but with any armor which the user has.
func (skin *McSkin) renderUpperArmor() *image.NRGBA {
	// This will be the base.
	upperArmorBodyImg := skin.renderUpperBody()

	// If it's an old skin, they don't have armor here.
	if skin.is18Skin() {
		// Get the armor layers from the skin and remove the Alpha.
		torso2Img := imaging.Crop(skin.Image, image.Rect(Torso2X, Torso2Y, Torso2X+TorsoWidth, Torso2Y+TorsoHeight))
		skin.removeAlpha(torso2Img)

		la2Img := imaging.Crop(skin.Image, image.Rect(La2X, La2Y, La2X+LaWidth, La2Y+TorsoHeight))
		skin.removeAlpha(la2Img)

		ra2Img := imaging.Crop(skin.Image, image.Rect(Ra2X, Ra2Y, Ra2X+RaWidth, Ra2Y+TorsoHeight))
		skin.removeAlpha(ra2Img)

		return skin.drawUpper(upperArmorBodyImg, torso2Img, ra2Img, la2Img)
	}
	return upperArmorBodyImg
}

// Given a base, torso and arms, it will return them all arranged correctly.
func (skin *McSkin) drawUpper(base, torso, la, ra *image.NRGBA) *image.NRGBA {
	// Torso
	fastDraw(base, torso, LaWidth, 0)
	// Left Arm
	fastDraw(base, la, 0, 0)
	// Right Arm
	fastDraw(base, ra, LaWidth+TorsoWidth, 0)

	return base
}

// Returns the legs.
func (skin *McSkin) renderLowerBody() *image.NRGBA {
	// This will be the base.
	lowerBodyImg := image.NewNRGBA(image.Rect(0, 0, LlWidth+RlWidth, LlHeight))

	rlImg := imaging.Crop(skin.Image, image.Rect(RlX, RlY, RlX+RlWidth, RlY+RlHeight))

	// If it's an old skin, they don't have a Left Leg, so we'll just flip their right.
	var llImg image.Image
	if skin.is18Skin() {
		llImg = imaging.Crop(skin.Image, image.Rect(LlX, LlY, LlX+LlWidth, LlY+LlHeight))
	} else {
		llImg = imaging.FlipH(rlImg)
	}

	return skin.drawLower(lowerBodyImg, rlImg, llImg.(*image.NRGBA))
}

// Returns the legs but with any armor which the user has.
func (skin *McSkin) renderLowerArmor() *image.NRGBA {
	// This will be the base.
	lowerArmorBodyImg := skin.renderLowerBody()

	// If it's an old skin, they don't have armor here.
	if skin.is18Skin() {
		// Get the armor layers from the skin and remove the Alpha.
		ll2Img := imaging.Crop(skin.Image, image.Rect(Ll2X, Ll2Y, Ll2X+LlWidth, Ll2Y+LlHeight))
		skin.removeAlpha(ll2Img)

		rl2Img := imaging.Crop(skin.Image, image.Rect(Rl2X, Rl2Y, Rl2X+RlWidth, Rl2Y+RlHeight))
		skin.removeAlpha(rl2Img)

		return skin.drawLower(lowerArmorBodyImg, rl2Img, ll2Img)
	}
	return lowerArmorBodyImg
}

// Given a base and legs, it will return them all arranged correctly.
func (skin *McSkin) drawLower(base, ll, rl *image.NRGBA) *image.NRGBA {
	// Left Leg
	fastDraw(base, ll, 0, 0)
	// Right Leg
	fastDraw(base, rl, LlWidth, 0)

	return base
}

// Rams the head onto the base (hopefully body...) to return a Frankenstein.
func (skin *McSkin) addHead(base, head *image.NRGBA) *image.NRGBA {
	base.Pix = append(make([]uint8, HeadHeight*base.Stride), base.Pix...)
	base.Rect.Max.Y += HeadHeight
	fastDraw(base, head, LaWidth, 0)

	return base
}

// Attached the legs onto the base (likely body).
func (skin *McSkin) addLegs(base, legs *image.NRGBA) *image.NRGBA {
	base.Pix = append(base.Pix, make([]uint8, LlHeight*base.Stride)...)
	base.Rect.Max.Y += LlHeight
	fastDraw(base, legs, LaWidth, HeadHeight+TorsoHeight)

	return base
}

// Writes the *processed* image as a PNG to the given writer.
func (skin *McSkin) WritePNG(w io.Writer) error {
	return png.Encode(w, skin.Processed)
}

// Writes the processed image as an svg.
func (skin *McSkin) WriteSVG(w io.Writer) error {
	canvas := svg.New(w)
	bounds := skin.Processed.Bounds()
	img := skin.Processed.(*image.NRGBA)

	// Make a canvas the same size as the image.
	canvas.Start(bounds.Max.X, bounds.Max.Y, `shape-rendering="crispEdges"`)
	// Loop through every pixel in the image.
	for y := 0; y < bounds.Max.Y; y += 1 {
		for x := 0; x < bounds.Max.X; x += 1 {
			ptr := y*img.Stride + x*4

			// Only draw opaque pixels.
			if img.Pix[ptr+3] == 0xFF {
				canvas.Rect(x, y, 1, 1, "fill:rgb("+
					strconv.Itoa(int(img.Pix[ptr]))+","+
					strconv.Itoa(int(img.Pix[ptr+1]))+","+
					strconv.Itoa(int(img.Pix[ptr+2]))+")")
			}

		}
	}
	canvas.End()
	return nil
}

// Writes the *original* skin image as a png to the given writer.
func (skin *McSkin) WriteSkin(w io.Writer) error {
	return png.Encode(w, skin.Image)
}

// Resizes the skin to the given dimensions, keeping aspect ratio.
func (skin *McSkin) resize(filter imaging.ResampleFilter) {
	if skin.Type != ImageTypeSVG {
		skin.Processed = imaging.Resize(skin.Processed, skin.Width, 0, filter)
	}
}

// Removes the skin's alpha matte from the given image.
func (skin *McSkin) removeAlpha(img *image.NRGBA) {
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
func (skin *McSkin) is18Skin() bool {
	bounds := skin.Image.Bounds()
	return bounds.Max.Y == 64
}

// Returns the head of the skin image.
func (skin *McSkin) cropHead(img image.Image) image.Image {
	return imaging.Crop(img, image.Rect(HeadX, HeadY, HeadX+HeadWidth, HeadY+HeadHeight))
}

// Returns the head of the skin image overlayed with the helm.
func (skin *McSkin) cropHelm(img image.Image) image.Image {
	headImg := skin.cropHead(img)
	helmImg := imaging.Crop(img, image.Rect(HelmX, HelmY, HelmX+HeadWidth, HelmY+HeadHeight))
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

	// Consider switch to uints https://github.com/minotar/imgd/pull/214#discussion_r869854923
	srcColor := [4]float32{}
	dstColor := [4]float32{}
	finalColor := [4]uint8{}

	pointer := dst.PixOffset(x, y)
	for row := 0; row < maxY; row += 1 {
		for i := 0; i < maxX; i += 4 {
			srcPx := row*src.Stride + i
			dstPx := row*dst.Stride + i + pointer

			uintToFloat(src, srcPx, &srcColor)
			uintToFloat(dst, dstPx, &dstColor)

			// Blend
			sR := srcColor[0]
			sG := srcColor[1]
			sB := srcColor[2]
			sA := srcColor[3]

			dR := dstColor[0]
			dG := dstColor[1]
			dB := dstColor[2]
			// dA := dstColor[3]

			dstColor[0] = (sR * sA) + (dR * (1.0 - sA))
			dstColor[1] = (sG * sA) + (dG * (1.0 - sA))
			dstColor[2] = (sB * sA) + (dB * (1.0 - sA))
			dstColor[3] = 1.0 // Alpha Func (sA * 1.0) + (dA * 0.0)

			floatToUint(&dstColor, &finalColor)
			dst.Pix[dstPx+0] = finalColor[0]
			dst.Pix[dstPx+1] = finalColor[1]
			dst.Pix[dstPx+2] = finalColor[2]
			dst.Pix[dstPx+3] = finalColor[3]

		}
	}
}

func uintToFloat(data *image.NRGBA, offset int, color *[4]float32) {
	color[0] = float32(data.Pix[offset]) / 255.0
	color[1] = float32(data.Pix[offset+1]) / 255.0
	color[2] = float32(data.Pix[offset+2]) / 255.0
	color[3] = float32(data.Pix[offset+3]) / 255.0
}

func floatToUint(color *[4]float32, final *[4]uint8) {
	final[0] = uint8(color[0] * 255)
	final[1] = uint8(color[1] * 255)
	final[2] = uint8(color[2] * 255)
	final[3] = uint8(color[3] * 255)
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
