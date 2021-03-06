package defparse

import (
	"image"
	"image/color"
	"image/draw"
)

var (
	background            = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	shadowBorder          = color.RGBA{R: 255, G: 150, B: 255, A: 255}
	shadowBody            = color.RGBA{R: 255, G: 0, B: 255, A: 255}
	selection             = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	selectionShadowBody   = color.RGBA{R: 180, G: 0, B: 255, A: 255}
	selectionShadowBorder = color.RGBA{R: 0, G: 255, B: 0, A: 255}
)

func decodePixels(pixels []uint8, palette color.Palette, imgMeta *ImageMeta) *image.RGBA {
	originImgMax := image.Pt(int(imgMeta.Width), int(imgMeta.Height))
	originImgRect := image.Rectangle{Min: image.Pt(0, 0), Max: originImgMax}
	img := image.NewPaletted(originImgRect, palette)
	img.Pix = pixels

	imgRGBA := image.NewRGBA(image.Rect(0, 0, int(imgMeta.FullWight), int(imgMeta.FullHeight)))
	margin := image.Pt(int(imgMeta.LeftMargin), int(imgMeta.TopMargin))
	innerRect := image.Rectangle{Min: margin, Max: margin.Add(originImgMax)}
	draw.Draw(imgRGBA, innerRect, img, image.Pt(0, 0), draw.Src)

	replaceDefSpecialColors(imgRGBA, imgMeta)
	return imgRGBA
}

func replaceDefSpecialColors(img *image.RGBA, imgMeta *ImageMeta) {
	for x := int(imgMeta.LeftMargin); x < int(imgMeta.LeftMargin)+int(imgMeta.Width); x++ {
		for y := int(imgMeta.TopMargin); y < int(imgMeta.TopMargin)+int(imgMeta.Height); y++ {
			if img.At(x, y) == background {
				img.Set(x, y, color.RGBA{A: 0})
			}
			if img.At(x, y) == shadowBorder {
				img.Set(x, y, color.RGBA{A: 64})
			}
			if img.At(x, y) == shadowBody {
				img.Set(x, y, color.RGBA{A: 128})
			}
			//if img.At(x, y) == selection {
			//	img.Set(x, y, color.RGBA{A: 0})
			//}
			if img.At(x, y) == selectionShadowBody {
				img.Set(x, y, color.RGBA{A: 128})
			}
			if img.At(x, y) == selectionShadowBorder {
				img.Set(x, y, color.RGBA{A: 64})
			}
		}
	}
}
