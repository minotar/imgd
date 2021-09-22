package minecraft

import (
	"bytes"
	"encoding/base64"
	"image"

	// If we work with PNGs we need this
	_ "image/png"

	"github.com/pkg/errors"
)

const SteveHash = "98903c1609352e11552dca79eb1ce3d6"

func (s *Skin) FetchSteve() error {
	bytes, err := GetSteveBytes()
	if err != nil {
		return err
	}

	err = s.Decode(bytes)
	if err != nil {
		return errors.Wrap(err, "failed to decode Steve skin")
	}

	return nil
}

func GetSteveBytes() (*bytes.Buffer, error) {
	steveImgBytes, err := base64.StdEncoding.DecodeString(SteveBase64)
	if err != nil {
		return bytes.NewBuffer([]byte{}), errors.Wrap(err, "failed to GetSteveBytes")
	}

	return bytes.NewBuffer(steveImgBytes), nil
}

func FetchImageForSteve() (image.Image, error) {
	bytes, err := GetSteveBytes()
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode Steve image")
	}
	return img, nil
}

func FetchSkinForSteve() (Skin, error) {
	skin := &Skin{}

	return *skin, skin.FetchSteve()
}

// The constant below contains Mojang AB copyrighted content.
// The use of this imagery is subject to their copyright.

const SteveBase64 = "iVBORw0KGgoAAAANSUhEUgAAAEAAAAAgCAYAAACinX6EAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyhpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMDE0IDc5LjE1Njc5NywgMjAxNC8wOC8yMC0wOTo1MzowMiAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIDIwMTQgKE1hY2ludG9zaCkiIHhtcE1NOkluc3RhbmNlSUQ9InhtcC5paWQ6NEM1NEZBNzk3OTExMTFFNDgwQUVDNTg1RjUyMzAwMTMiIHhtcE1NOkRvY3VtZW50SUQ9InhtcC5kaWQ6NEM1NEZBN0E3OTExMTFFNDgwQUVDNTg1RjUyMzAwMTMiPiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDo0QzU0RkE3Nzc5MTExMUU0ODBBRUM1ODVGNTIzMDAxMyIgc3RSZWY6ZG9jdW1lbnRJRD0ieG1wLmRpZDo0QzU0RkE3ODc5MTExMUU0ODBBRUM1ODVGNTIzMDAxMyIvPiA8L3JkZjpEZXNjcmlwdGlvbj4gPC9yZGY6UkRGPiA8L3g6eG1wbWV0YT4gPD94cGFja2V0IGVuZD0iciI/PgRvamEAAAXnSURBVHja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"
