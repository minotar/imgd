package mcskin_test

import (
	"os"
	"testing"

	"github.com/minotar/imgd/pkg/mcclient/mcuser"
	"github.com/minotar/imgd/pkg/minecraft"
	"github.com/minotar/imgd/pkg/processd/mcskin"
	"github.com/minotar/imgd/pkg/util/sample_skin"
)

func getMcSkin() (*mcskin.McSkin, error) {
	textureRC, err := sample_skin.GetSampleSkinReadCloser()
	if err != nil {
		return nil, err
	}
	textureIO := mcuser.TextureIO{ReadCloser: textureRC}
	texture, err := textureIO.DecodeTexture()
	if err != nil {
		return nil, err
	}
	return &mcskin.McSkin{
		Skin:  minecraft.Skin{Texture: texture},
		Type:  mcskin.ImageTypePNG,
		Width: 180,
	}, nil
}

func writeMcSkin(mcSkin *mcskin.McSkin, filename string) error {
	filePNG, err := os.Create(filename + ".png")
	if err != nil {
		return err
	}
	defer filePNG.Close()

	err = mcSkin.WritePNG(filePNG)
	if err != nil {
		return err
	}

	fileSVG, err := os.Create(filename + ".svg")
	if err != nil {
		return err
	}
	defer fileSVG.Close()
	return mcSkin.WriteSVG(fileSVG)
}

func TestRenderBody(t *testing.T) {
	mcSkin, err := getMcSkin()
	if err != nil {
		t.Fatalf("Unable to get mcSkin: %s", err)
	}
	mcSkin.GetBody()
	writeMcSkin(mcSkin, "test_render_body")
}

func TestRenderArmorBody(t *testing.T) {
	mcSkin, err := getMcSkin()
	if err != nil {
		t.Fatalf("Unable to get mcSkin: %s", err)
	}
	mcSkin.GetArmorBody()
	writeMcSkin(mcSkin, "test_render_armor_body")
}

func BenchmarkRenderArmorBody(b *testing.B) {
	mcSkin, err := getMcSkin()
	if err != nil {
		b.Fatalf("Unable to get mcSkin: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mcSkin.GetArmorBody()
		mcSkin.Processed = nil
	}
}
