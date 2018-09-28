/*
	go-osk - Simple on-screen keyboard aimed at Kobo ereaders
    Copyright (C) 2018 Sherman Perry

    This file is part of go-osk.

    go-osk is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    go-osk is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with go-osk.  If not, see <https://www.gnu.org/licenses/>.
*/

package osk

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

const (
	KTstandardChar   = 0
	KTcarriageReturn = 1
	KTbackspace      = 2
	KTdelete         = 3
	KTcapsLock       = 4
	KTcontrol        = 5
	KTalt            = 6
)

type KeyMap struct {
	Lang      string `json:"lang"`
	KBmargins struct {
		Top    float64 `json:"top"`
		Bottom float64 `json:"bottom"`
		Left   float64 `json:"left"`
		Right  float64 `json:"right"`
	} `json:"kbMargins"`
	TotalKeyWidth  float64 `json:"totalKeyWidth"`
	TotalRowHeight float64 `json:"totalRowHeight"`
	Rows           []struct {
		RowHeight float64 `json:"rowHeight"`
		Keys      []struct {
			IsPadding bool    `json:"isPadding"`
			KeyType   int     `json:"keyType"`
			KeyWidth  float64 `json:"keyWidth"`
			Char      string  `json:"char"`
		} `json:"keys"`
	} `json:"rows"`
}

type coords struct {
	X int
	Y int
}

// Key contains information about each key on the virtual keyboard
type Key struct {
	coord   coords
	width   int
	IsKey   bool
	KeyType int
	KeyCode rune
}

// Row contains the height, and all the keys in the row.
type Row struct {
	rowHeight int
	keys      []Key
}

// VirtKeyboard contains the entire virtual keyboard, and the methods
// required to generate one from a keymap file
type VirtKeyboard struct {
	widthPX         int
	heightPX        int
	kmUnitWidth     int
	rhUnitWidth     int
	StartCoords     coords
	rows            []Row
	debounceStartTm time.Time
	prevKey         Key
}

// validateKeymap checks that the keymap file contains valid measurements
func validateKeymap(km *KeyMap) error {
	// Margin check
	if km.KBmargins.Bottom < 0 ||
		km.KBmargins.Top < 0 ||
		km.KBmargins.Left < 0 ||
		km.KBmargins.Right < 0 {
		return errors.New("keymap: negative numbers not allowed")
	} else if (km.KBmargins.Top+km.KBmargins.Bottom) > 0.8 &&
		(km.KBmargins.Left+km.KBmargins.Right) > 0.8 {
		return errors.New("combined margins exceed 0.8")
	}
	// Measurement check
	currRowMeas := float64(0)
	for i, r := range km.Rows {
		currRowMeas += r.RowHeight
		currKeyMeas := float64(0)
		for _, k := range r.Keys {
			currKeyMeas += k.KeyWidth
		}
		if currKeyMeas > km.TotalKeyWidth {
			errMsg := fmt.Sprintf("Key widths sum exceeds %f in row %d", km.TotalKeyWidth, i)
			return errors.New(errMsg)
		}
	}
	if currRowMeas > km.TotalRowHeight {
		errMsg := fmt.Sprintf("Row heights sum exceeds %f", km.TotalRowHeight)
		return errors.New(errMsg)
	}
	return nil
}

// New initilizes a VirtKeyboard for use
func New(km *KeyMap, fbWidth, fbHeight int) (*VirtKeyboard, error) {
	v := &VirtKeyboard{}
	if err := validateKeymap(km); err != nil {
		return v, err
	}
	// Calculate our margins in px from the percentages provided in the keymap
	floatFBw, floatFBh := float64(fbWidth), float64(fbHeight)
	pxFromTop := int(math.Round(floatFBh * km.KBmargins.Top))
	pxFromBot := int(math.Round(floatFBh * km.KBmargins.Bottom))
	pxFromLeft := int(math.Round(floatFBw * km.KBmargins.Left))
	pxFromRight := int(math.Round(floatFBw * km.KBmargins.Right))

	// Calculate our origin and dimensions from the margins
	v.StartCoords.X, v.StartCoords.Y = pxFromLeft, pxFromTop
	v.widthPX = fbWidth - pxFromLeft - pxFromRight
	v.heightPX = fbHeight - pxFromTop - pxFromBot

	// What's the width of each keymap unit? Rounded down to the nearest pixel of course
	v.kmUnitWidth = int(float64(v.widthPX) / km.TotalKeyWidth)
	// And the height of each rowheight unit?
	v.rhUnitWidth = int(float64(v.heightPX) / km.TotalRowHeight)

	// time to give our keymap into a set of usable coordinates
	v.convertKeymap(km)
	return v, nil
}

// convertKeymap converts a keymap file into rows of keys with coordinate
// information
func (v *VirtKeyboard) convertKeymap(km *KeyMap) {
	currY := v.StartCoords.Y
	for _, r := range km.Rows {
		row := Row{}
		row.rowHeight = int(float64(v.rhUnitWidth) * r.RowHeight)
		ky := make([]Key, len(r.Keys))
		currX := v.StartCoords.X
		for j, k := range r.Keys {
			ky[j].width = int(float64(v.kmUnitWidth) * k.KeyWidth)
			ky[j].coord.Y = currY
			ky[j].coord.X = currX
			currX += ky[j].width
			ky[j].KeyType = k.KeyType
			ky[j].IsKey = !k.IsPadding
			if ky[j].KeyType == 0 && len(k.Char) > 0 {
				runeSlice := []rune(k.Char)
				// We only care about the first rune...
				ky[j].KeyCode = runeSlice[0]
			} else {
				ky[j].KeyCode = 0
			}
		}
		row.keys = ky
		currY += row.rowHeight
		v.rows = append(v.rows, row)
	}
}

// GetLabel returns a label for "special" keys
func (v *VirtKeyboard) GetLabel(kt int) string {
	switch kt {
	case KTalt:
		return "ALT"
	case KTbackspace:
		return "BKSP"
	case KTcapsLock:
		return "CPLK"
	case KTcarriageReturn:
		return "RET"
	case KTcontrol:
		return "CTRL"
	case KTdelete:
		return "DEL"
	}
	return ""
}

// CreateIMG generates an image from the current keyboard.CreateIMG
// The current implementation saves the image as a PNG. This behaviour may
// change in the future to return an RBGA image
func (v *VirtKeyboard) CreateIMG(savePath, fontPath string) {
	kc := gg.NewContext(v.widthPX, v.heightPX)
	kc.DrawRectangle(0, 0, float64(v.widthPX), float64(v.heightPX))
	kc.SetRGB255(240, 240, 240)
	kc.Fill()
	kc.LoadFontFace(fontPath, 36)
	for _, r := range v.rows {
		for _, k := range r.keys {
			if k.IsKey {
				kc.Push()
				kx, ky := float64(k.coord.X-v.StartCoords.X), float64((k.coord.Y - v.StartCoords.Y))
				kw, kh := float64(k.width), float64(r.rowHeight)
				kmx, kmy := (kx + kw/2), (ky + kh/2)
				kc.DrawRectangle(kx, ky, kw, kh)
				kc.SetRGB255(0, 0, 0)
				kc.StrokePreserve()
				kc.SetRGB255(255, 255, 255)
				kc.Fill()
				kc.SetRGB255(0, 0, 0)
				if k.KeyType == KTstandardChar {
					kc.DrawStringAnchored(strings.ToUpper(string(k.KeyCode)), kmx, kmy, 0.5, 0.5)
				} else {
					kc.DrawStringAnchored(v.GetLabel(k.KeyType), kmx, kmy, 0.5, 0.5)
				}
				kc.Pop()
			}
		}
	}
	kc.SavePNG(savePath)
}

// GetPressedKey uses the coordinates provided to determine which
// key was pressed, and returns the key
func (v *VirtKeyboard) GetPressedKey(inX, inY int) (Key, error) {
	// Given X and Y, we need to find which key was pressed

	// First, reject any coordinates that are out of bounds
	if inY < v.StartCoords.Y || inY > (v.StartCoords.Y+v.heightPX) {
		return Key{}, errors.New("Y out of bounds")
	} else if inX < v.StartCoords.X || inX > (v.StartCoords.X+v.widthPX) {
		return Key{}, errors.New("Y out of bounds")
	}
	// Get the row index.
	rowIndex := -1
	currY := v.StartCoords.Y
	for i, r := range v.rows {
		if inY <= currY+r.rowHeight {
			rowIndex = i
			break
		}
		currY += r.rowHeight
	}
	// Getting key in row is a little trickier, as key width varies
	// Linear search, because our list will never be very large...
	if rowIndex >= 0 {
		keyNum := len(v.rows[rowIndex].keys)
		for i := 0; i < keyNum; i++ {
			k := v.rows[rowIndex].keys[i]
			if inX <= (k.coord.X + k.width) {
				if k != v.prevKey {
					v.prevKey = k
					v.debounceStartTm = time.Now()
					return k, nil
				} else {
					if time.Since(v.debounceStartTm) < (50 * time.Millisecond) {
						return Key{}, errors.New("debounce detected")
					} else {
						v.prevKey = k
						v.debounceStartTm = time.Now()
						return k, nil
					}
				}
			}
		}
	}
	return Key{}, errors.New("key not found")
}
