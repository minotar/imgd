package render

import (
	"github.com/minotar/imgd/pkg/mcskin"
	"github.com/minotar/imgd/pkg/skin"
)

var (
	DefaultProcessRoutes = map[string]skin.SkinProcessor{
		"Avatar":                 mcskin.HandlerHead,
		"Helm":                   mcskin.HandlerHelm,
		"Cube":                   mcskin.HandlerCube,
		"CubeHelm":               mcskin.HandlerCubeHelm,
		"Bust":                   mcskin.HandlerBust,
		"Body":                   mcskin.HandlerBody,
		"Armor/Bust|Armour/Bust": mcskin.HandlerArmorBust,
		"Armor/Body|Armour/Body": mcskin.HandlerArmorBody,
	}
)
