// textures_test.go
package minecraft

import (
	"testing"

	"github.com/minotar/imgd/pkg/minecraft/mockminecraft"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTextures(t *testing.T) {

	Convey("Test Texture.fetch", t, func() {

		Convey("clone1018 texture should return the correct skin", func() {
			texture := &Texture{Mc: mcTest, URL: "http://textures.minecraft.net/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649"}

			err := texture.Fetch()

			So(err, ShouldBeNil)
			So(texture.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")
		})

		Convey("Bad texture requests should gracefully fail", func() {

			Convey("Bad texture URL (invalid-image)", func() {
				texture := &Texture{Mc: mcTest, URL: mockminecraft.TestURL + "/texture/MalformedTexture"}

				err := texture.Fetch()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unable to Decode Texture: unable to CastToNRGBA: png: invalid format: not enough pixel data")
			})

			Convey("Bad texture URL (non-image)", func() {
				texture := &Texture{Mc: mcTest, URL: mockminecraft.TestURL + "/200"}

				err := texture.Fetch()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unable to Decode Texture: unable to CastToNRGBA: image: unknown format")
			})

			Convey("Bad texture URL (non-200)", func() {
				texture := &Texture{Mc: mcTest, URL: mockminecraft.TestURL + "/404"}

				err := texture.Fetch()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unable to Fetch Texture: apiRequest HTTP 404 Not Found")
			})

		})

	})

	Convey("Test DecodeTextureProperty + Texture.FetchWithTextureProperty", t, func() {

		Convey("Should correctly decode and fetch Skin and Cape URL", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("48a0a7e4d5594873a617dc189f76a8a1")
			profileTextureProperty, err1 := DecodeTextureProperty(sessionProfile)

			So(err1, ShouldBeNil)
			So(profileTextureProperty.Textures.Skin.URL, ShouldEqual, "http://textures.minecraft.net/texture/e1c6c9b6de88f4188f9732909c76dfcd7b16a40a031ce1b4868e4d1f8898e4f")
			So(profileTextureProperty.Textures.Cape.URL, ShouldEqual, "http://textures.minecraft.net/texture/c3af7fb821254664558f28361158ca73303c9a85e96e5251102958d7ed60c4a3")

			skin := &Texture{Mc: mcTest}
			err2 := skin.FetchWithTextureProperty(profileTextureProperty, "Skin")
			So(err2, ShouldBeNil)
			So(skin.Hash, ShouldEqual, "c05454f331fa93b3e38866a9ec52c467")

			cape := &Texture{Mc: mcTest}
			err3 := cape.FetchWithTextureProperty(profileTextureProperty, "Cape")
			So(err3, ShouldBeNil)
			So(cape.Hash, ShouldEqual, "8cbf8786caba2f05383cf887be592ee6")

		})

		Convey("Should only decode and fetch Skin URL", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("d9135e082f2244c89cb0bee234155292")
			profileTextureProperty, err1 := DecodeTextureProperty(sessionProfile)

			So(err1, ShouldBeNil)
			So(profileTextureProperty.Textures.Skin.URL, ShouldEqual, "http://textures.minecraft.net/texture/cd9ca55e9862f003ebfa1872a9244ad5f721d6b9e6883dd1d42f87dae127649")
			So(profileTextureProperty.Textures.Cape.URL, ShouldBeBlank)

			skin := &Texture{Mc: mcTest}
			err2 := skin.FetchWithTextureProperty(profileTextureProperty, "Skin")
			So(err2, ShouldBeNil)
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")

			cape := &Texture{Mc: mcTest}
			err3 := cape.FetchWithTextureProperty(profileTextureProperty, "Cape")
			So(err3, ShouldNotBeNil)
			So(err3.Error(), ShouldEqual, "Cape URL not present")
		})

		Convey("Should error about fetching a malformed Skin texture", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000008")
			profileTextureProperty, err1 := DecodeTextureProperty(sessionProfile)

			So(err1, ShouldBeNil)
			So(profileTextureProperty.Textures.Skin.URL, ShouldEqual, "http://textures.minecraft.net/texture/MalformedTexture")
			So(profileTextureProperty.Textures.Cape.URL, ShouldBeBlank)

			skin := &Texture{Mc: mcTest}
			err2 := skin.FetchWithTextureProperty(profileTextureProperty, "Skin")
			So(err2, ShouldNotBeNil)
			So(err2.Error(), ShouldEqual, "FetchWithTextureProperty failed: unable to Decode Texture: unable to CastToNRGBA: png: invalid format: not enough pixel data")
		})

		Convey("Should error about no textures", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000004")
			profileTextureProperty, err := DecodeTextureProperty(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to DecodeTextureProperty: no textures property")
			So(profileTextureProperty, ShouldResemble, SessionProfileTextureProperty{})
		})

		Convey("Should error trying to decode", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000005")
			profileTextureProperty, err := DecodeTextureProperty(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to DecodeTextureProperty: unexpected EOF")
			So(profileTextureProperty, ShouldResemble, SessionProfileTextureProperty{})
		})

		Convey("Should error trying to decode an unknown textureType", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("5c115ca73efd41178213a0aff8ef11e0")
			profileTextureProperty, _ := DecodeTextureProperty(sessionProfile)

			skin := &Texture{Mc: mcTest}
			err := skin.FetchWithTextureProperty(profileTextureProperty, "Bandana")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Unknown textureType")
		})

	})

	Convey("Test FetchWithSessionProfile", t, func() {

		Convey("Should correctly fetch Skin and Cape URL", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("48a0a7e4d5594873a617dc189f76a8a1")

			skin := &Texture{Mc: mcTest}
			err1 := skin.FetchWithSessionProfile(sessionProfile, "Skin")

			So(err1, ShouldBeNil)
			So(skin.Hash, ShouldEqual, "c05454f331fa93b3e38866a9ec52c467")

			cape := &Texture{Mc: mcTest}
			err2 := cape.FetchWithSessionProfile(sessionProfile, "Cape")

			So(err2, ShouldBeNil)
			So(cape.Hash, ShouldEqual, "8cbf8786caba2f05383cf887be592ee6")
		})

		Convey("Should only fetch Skin URL", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("d9135e082f2244c89cb0bee234155292")

			skin := &Texture{Mc: mcTest}
			err1 := skin.FetchWithSessionProfile(sessionProfile, "Skin")

			So(err1, ShouldBeNil)
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")

			cape := &Texture{Mc: mcTest}
			err2 := cape.FetchWithSessionProfile(sessionProfile, "Cape")

			So(err2, ShouldNotBeNil)
			So(err2.Error(), ShouldEqual, "FetchWithSessionProfile failed: Cape URL not present")
		})

		Convey("Should error about no textures", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000004")
			skin := &Texture{Mc: mcTest}
			err := skin.FetchWithSessionProfile(sessionProfile, "Skin")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to DecodeTextureProperty: no textures property")
		})

		Convey("Should error trying to decode", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000005")
			skin := &Texture{Mc: mcTest}
			err := skin.FetchWithSessionProfile(sessionProfile, "Skin")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to DecodeTextureProperty: unexpected EOF")
		})

	})

	Convey("Test FetchWithUsername", t, func() {

		Convey("Should correctly fetch Skin and Cape URL", func() {
			skin := &Texture{Mc: mcTest}
			err1 := skin.FetchWithUsername("citricsquid", "Skin")

			So(err1, ShouldBeNil)
			So(skin.Hash, ShouldEqual, "c05454f331fa93b3e38866a9ec52c467")

			cape := &Texture{Mc: mcTest}
			err2 := cape.FetchWithUsername("citricsquid", "Cape")

			So(err2, ShouldBeNil)
			So(cape.Hash, ShouldEqual, "8cbf8786caba2f05383cf887be592ee6")
		})

		Convey("Should only fetch Skin URL", func() {
			skin := &Texture{Mc: mcTest}
			err1 := skin.FetchWithUsername("clone1018", "Skin")

			So(err1, ShouldBeNil)
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")

			cape := &Texture{Mc: mcTest}
			err2 := cape.FetchWithUsername("clone1018", "Cape")

			So(err2, ShouldNotBeNil)
			So(err2.Error(), ShouldEqual, "FetchWithUsername failed: unable to Fetch Texture: apiRequest HTTP 404 Not Found")
		})

		Convey("Should error about fetching a malformed Skin texture", func() {
			skin := &Texture{Mc: mcTest}
			err := skin.FetchWithUsername("MalformedTexture", "Skin")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "FetchWithUsername failed: unable to Decode Texture: unable to CastToNRGBA: png: invalid format: not enough pixel data")
		})

		Convey("Should error trying to fetch the 404 skin", func() {
			skin := &Texture{Mc: mcTest}
			err := skin.FetchWithUsername("404STexture", "Skin")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "FetchWithUsername failed: unable to Fetch Texture: apiRequest HTTP 404 Not Found")
		})

		Convey("Should error trying to decode an unknown textureType", func() {
			skin := &Texture{Mc: mcTest}
			err := skin.FetchWithUsername("LukeHandle", "Bandana")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Unkown textureType or missing UsernameAPI lookup URL")
		})

		Convey("Should error with no UsernameAPI", func() {
			skin := &Texture{Mc: mcProd} // mcProd does not have the UsernameAPI set on it
			err := skin.FetchWithUsername("LukeHandle", "Skin")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Unkown textureType or missing UsernameAPI lookup URL")
		})

	})

	Convey("Test Steve", t, func() {

		Convey("Steve should return valid image", func() {
			steveImg, err := FetchImageForSteve()

			So(err, ShouldBeNil)
			So(steveImg, ShouldNotBeNil)
		})

		Convey("Steve should return valid skin", func() {
			steveSkin, err := FetchSkinForSteve()

			So(err, ShouldBeNil)
			So(steveSkin, ShouldNotResemble, Skin{Texture{Mc: mcTest}})
			So(steveSkin.Hash, ShouldEqual, "98903c1609352e11552dca79eb1ce3d6")
		})

	})

	Convey("Test Skins", t, func() {

		Convey("clone1018 should return valid image from Mojang", func() {
			skin, err := mcTest.FetchSkinUsername("clone1018")

			So(err, ShouldBeNil)
			So(skin, ShouldNotResemble, Skin{Texture{Mc: mcTest}})
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")
		})

		Convey("d9135e082f2244c89cb0bee234155292 should return valid image from Mojang", func() {
			skin, err := mcTest.FetchSkinUUID("d9135e082f2244c89cb0bee234155292")

			So(err, ShouldBeNil)
			So(skin, ShouldNotResemble, Skin{Texture{Mc: mcTest}})
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")
		})

		Convey("10000000000000000000000000000000 err from Mojang", func() {
			skin, err := mcTest.FetchSkinUUID("10000000000000000000000000000000")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to GetSessionProfile: user not found")
			So(skin, ShouldResemble, Skin{Texture{Mc: mcTest}})
		})

	})

	Convey("Test Capes", t, func() {

		Convey("citricsquid should return a Cape from Mojang", func() {
			cape, err := mcTest.FetchCapeUsername("citricsquid")

			So(err, ShouldBeNil)
			So(cape, ShouldNotResemble, Cape{Texture{Mc: mcTest}})
			So(cape.Hash, ShouldEqual, "8cbf8786caba2f05383cf887be592ee6")
		})

		Convey("48a0a7e4d5594873a617dc189f76a8a1 should return a Cape from Mojang", func() {
			cape, err := mcTest.FetchCapeUUID("48a0a7e4d5594873a617dc189f76a8a1")

			So(err, ShouldBeNil)
			So(cape, ShouldNotResemble, Cape{Texture{Mc: mcTest}})
			So(cape.Hash, ShouldEqual, "8cbf8786caba2f05383cf887be592ee6")
		})

		Convey("2f3665cc5e29439bbd14cb6d3a6313a7 should err from Mojang (No Cape)", func() {
			// lukegb
			cape, err := mcTest.FetchCapeUUID("2f3665cc5e29439bbd14cb6d3a6313a7")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "FetchWithSessionProfile failed: Cape URL not present")
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
			So(cape.Hash, ShouldBeBlank)
		})

		Convey("10000000000000000000000000000000 should err from Mojang (No User)", func() {
			cape, err := mcTest.FetchCapeUUID("10000000000000000000000000000000")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to GetSessionProfile: user not found")
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
		})

	})

	// This could be a lot more DRY but shush
	Convey("Test FetchTexturesWithSessionProfile", t, func() {

		Convey("clone1018", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("d9135e082f2244c89cb0bee234155292")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not able to retrieve cape: Cape URL not present")
			So(user.Username, ShouldEqual, "clone1018")
			So(user.UUID, ShouldEqual, "d9135e082f2244c89cb0bee234155292")
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")
		})

		Convey("citricsquid", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("48a0a7e4d5594873a617dc189f76a8a1")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldBeNil)
			So(user.Username, ShouldEqual, "citricsquid")
			So(user.UUID, ShouldEqual, "48a0a7e4d5594873a617dc189f76a8a1")
			So(cape.Hash, ShouldEqual, "8cbf8786caba2f05383cf887be592ee6")
			So(skin.Hash, ShouldEqual, "c05454f331fa93b3e38866a9ec52c467")
		})

		Convey("NoTexture", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000004")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to decode sessionProfile: unable to DecodeTextureProperty: no textures property")
			So(user.Username, ShouldEqual, "NoTexture")
			So(skin, ShouldResemble, Skin{Texture{Mc: mcTest}})
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
		})

		Convey("MalformedTexProp", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000005")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to decode sessionProfile: unable to DecodeTextureProperty: unexpected EOF")
			So(user.Username, ShouldEqual, "MalformedTexProp")
			So(skin, ShouldResemble, Skin{Texture{Mc: mcTest}})
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
		})

		Convey("MalformedSTex", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000008")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not able to retrieve skin: FetchWithTextureProperty failed: unable to Decode Texture: unable to CastToNRGBA: png: invalid format: not enough pixel data")
			So(user.Username, ShouldEqual, "MalformedSTex")
			So(skin, ShouldResemble, Skin{Texture{Mc: mcTest, Source: "SessionProfile", URL: "http://textures.minecraft.net/texture/MalformedTexture"}})
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
		})

		Convey("MalformedCTex", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000009")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not able to retrieve cape: FetchWithTextureProperty failed: unable to Decode Texture: unable to CastToNRGBA: png: invalid format: not enough pixel data")
			So(user.Username, ShouldEqual, "MalformedCTex")
			So(skin.Source, ShouldEqual, "SessionProfile")
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest, Source: "SessionProfile", URL: "http://textures.minecraft.net/texture/MalformedTexture"}})
		})

		Convey("404STexture", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000010")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not able to retrieve skin: FetchWithTextureProperty failed: unable to Fetch Texture: apiRequest HTTP 404 Not Found")
			So(user.Username, ShouldEqual, "404STexture")
			So(skin, ShouldResemble, Skin{Texture{Mc: mcTest, Source: "SessionProfile", URL: "http://textures.minecraft.net/texture/404Texture"}})
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest}})
		})

		Convey("404CTexture", func() {
			sessionProfile, _ := mcTest.GetSessionProfile("00000000000000000000000000000011")
			user, skin, cape, err := mcTest.FetchTexturesWithSessionProfile(sessionProfile)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "not able to retrieve cape: FetchWithTextureProperty failed: unable to Fetch Texture: apiRequest HTTP 404 Not Found")
			So(user.Username, ShouldEqual, "404CTexture")
			So(cape, ShouldResemble, Cape{Texture{Mc: mcTest, Source: "SessionProfile", URL: "http://textures.minecraft.net/texture/404Texture"}})
			So(skin.Source, ShouldEqual, "SessionProfile")
			So(skin.Hash, ShouldEqual, "a04a26d10218668a632e419ab073cf57")
		})

	})

}
