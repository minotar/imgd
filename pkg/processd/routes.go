package processd

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minotar/imgd/pkg/util/route_helpers"
	"github.com/minotar/minecraft"
)

// Set the default, min and max width to resize processed images to
const (
	DefaultWidth = int(180)
	MinWidth     = int(8)
	MaxWidth     = int(300)
)

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

func GetHead(skin minecraft.Skin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		width := GetWidth(vars["width"])

		workingSkin := mcSkin{Skin: skin}
		workingSkin.GetHead(width)
		workingSkin.WritePNG(w)
	})
}

type SkinProcessor func(minecraft.Skin) http.Handler

type SkinWrapper func(SkinProcessor) http.Handler

// skinWrapper would eg. be SkinLookupWrapper
func RegisterRoutes(m *mux.Router, skinWrapper SkinWrapper, processRoutes map[string]SkinProcessor) {

	for resource, processor := range processRoutes {
		mainPath := "/{resource:" + strings.ToLower(resource) + "}" + route_helpers.UserPath
		m.Path(mainPath).Handler(skinWrapper(processor)).Name(resource)
		m.Path(mainPath + "/{width:[0-9]+}").Handler(skinWrapper(processor)).Name(resource)
	}

}
