// Package moasics provides the main logic for generating a photomosaic from a set of smaller images
package mosaics

import (
	"fmt"
	"image"
	"image/draw"
)

// TileSize is the expected size of each subImage. If images are larger than this value, only the top-left corner up to TileSize will be used.
// If any dimension is smaller than this, they will be backfilled with white. If possible, subImages should be prescaled to a square of this size.
const DefaultTileSize = 90

//Maximum dimension in an finalized mosaic. Target image will be scaled up or down such that its largest side is this length.
const DefaultMaxDimension = DefaultTileSize * 70

// BuildMosaic builds a master image from the provided subimages.
func BuildMosaic(master image.Image, subImages []image.Image, evaluator Evaluator) image.Image {
	lib := NewLibrary(evaluator)
	for _, img := range subImages {
		lib.AddImage(img)
	}
	return BuildMosaicFromLibrary(master, lib)
}

func BuildMosaicFromLibrary(master image.Image, tiles *ThumbnailLibrary) image.Image {
	tileSize := DefaultTileSize
	dim := getMosaicDimensions(master.Bounds().Dx(), master.Bounds().Dy(), DefaultMaxDimension, tileSize)
	output := image.NewRGBA(image.Rect(0, 0, dim.width, dim.height))
	for tileY := 0; tileY < dim.tilesY; tileY++ {
		for tileX := 0; tileX < dim.tilesX; tileX++ {
			c := tiles.evaluator.Evaluate(master, tileX*dim.sourcePixelsPerTileX, tileY*dim.sourcePixelsPerTileY, dim.sourcePixelsPerTileX, dim.sourcePixelsPerTileY)
			tile := tiles.getBestMatch(c)
			rect := tile.Bounds().Add(image.Point{tileX * tileSize, tileY * tileSize})
			fmt.Println(rect)
			draw.Draw(output, rect, tile, image.ZP, draw.Over)
		}
	}
	return output
}

// Calclulates scaling factor for final mosaic so we can map original image tiles onto the final mosaic.
// This method values safety over perfect accuracy. We want to avoid overflowing the image bounds when iterating without needing
// to check. Because the minor dimension is stretched or truncated to be an even multiple of tilesize, we may see some distortion around the edges.
// This is ok. Mosaics are a bit fuzzy anyway, so we embrace this for convenience.
func getMosaicDimensions(originalX, originalY int, maxDimension, tileSize int) *mosaicDimensions {
	dim := mosaicDimensions{}
	if originalX >= originalY {
		dim.width = maxDimension
		dim.height = int(float64(originalY) * (float64(maxDimension) / float64(originalX)))
	} else {
		dim.height = maxDimension
		dim.width = int(float64(originalX) * (float64(maxDimension) / float64(originalY)))
	}
	// Make sure we are a multiple of tile size in both directions.
	if dim.height < tileSize {
		dim.height = tileSize
	}
	dim.height -= dim.height % tileSize
	if dim.width < tileSize {
		dim.width = tileSize
	}
	dim.width -= dim.width % tileSize

	//count tiles and source pixels per resultant tile
	dim.tilesX = dim.width / tileSize
	dim.tilesY = dim.height / tileSize
	dim.sourcePixelsPerTileX = originalX / dim.tilesX
	dim.sourcePixelsPerTileY = originalY / dim.tilesY
	return &dim
}

type mosaicDimensions struct {
	width                int
	height               int
	sourcePixelsPerTileX int
	sourcePixelsPerTileY int
	tilesX, tilesY       int
}
