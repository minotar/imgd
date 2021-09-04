package processd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minotar/minecraft"
)

func GetHead(skin minecraft.Skin) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workingSkin := mcSkin{Skin: skin}
		workingSkin.GetHead(64)
		workingSkin.WritePNG(w)
	})
}

// skinWrapper would eg. be SkinLookupWrapper
func RegisterRoutes(m *mux.Router, skinWrapper func(func(minecraft.Skin) http.Handler) http.Handler) {

	m.Path("/head/{user:" + minecraft.ValidUsernameRegex + "|" + minecraft.ValidUUIDPlainRegex + "}").Handler(skinWrapper(GetHead)).Name("GetHead")
}
