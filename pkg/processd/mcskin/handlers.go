// The problem that is sovled by using the handlers is that you can pass in the
// minecraft.Skin and it does the rest - vs. requiring a switch logic based on
// the route for which method to use.
package mcskin

import (
	"net/http"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/util/log"
)

// Will deliver an avatar Head when ServeHTTP  is called
func HandlerHead(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetHead
	return mcSkin.ServeHTTP
}

// Will deliver a head with Helm when ServeHTTP  is called
func HandlerHelm(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetHelm
	return mcSkin.ServeHTTP
}

// Will deliver an isometric "Cube" avatar when ServeHTTP  is called
func HandlerCube(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetCube
	return mcSkin.ServeHTTP
}

// Will deliver an isometric "Cube" with Helm when ServeHTTP  is called
func HandlerCubeHelm(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetCubeHelm
	return mcSkin.ServeHTTP
}

// Will deliver a Bust when ServeHTTP  is called
func HandlerBust(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetBust
	return mcSkin.ServeHTTP
}

// Will deliver a Bust with Armor when ServeHTTP  is called
func HandlerArmorBust(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetArmorBust
	return mcSkin.ServeHTTP
}

// Will deliver a Body when ServeHTTP  is called
func HandlerBody(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetBody
	return mcSkin.ServeHTTP
}

// Will deliver a Body with Armor when ServeHTTP  is called
func HandlerArmorBody(logger log.Logger, skinIO mcuser.TextureIO) http.HandlerFunc {
	mcSkin := &McSkin{Skin: skinIO.MustDecodeSkin(logger)}
	mcSkin.Processor = mcSkin.GetArmorBody
	return mcSkin.ServeHTTP
}

// The
func (skin *McSkin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if skin.Processor != nil {
		// If the Processor is set, use it to create the Processed image
		skin.Width, skin.Type = GetWidthType(r)
		skin.Processor()
	} else if skin.Processed == nil {
		// Otherwise, if there was no Processor and the Processed hadn't already
		// been set, then throw an error
		// This shouldn't happen outside development...
		//w.Header().Del("ETag")
		http.Error(w, "No skin processor or processed image to deliver", 500)
		return
	}

	switch skin.Type {
	case ImageTypePNG:
		w.Header().Add("Content-Type", string(ImageTypePNG))
		skin.WritePNG(w)
	case ImageTypeSVG:
		w.Header().Add("Content-Type", string(ImageTypeSVG))
		skin.WriteSVG(w)
	}
}
