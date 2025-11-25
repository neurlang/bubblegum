package lib

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
)

//go:embed fonts/*.png fonts/*.jpg
var embedFonts embed.FS

// Font represents a bitmap font loaded from PNG/JPEG files.
// It uses the same format as wayland/go-wayland-texteditor.
type Font struct {
	cellx   int
	celly   int
	mapping map[string][][3]byte
}

// hexfont is a simple 4x6 pixel font for rendering hex digits as placeholders
var hexfont = "" +
	" #    # ##  ##   #  ###  ## ###  #   #   #  ##   #  ##  ### ### ### " +
	"# #  ##   #   # #   #   #     # # # # # # # # # # # # # #   #   ### " +
	"# # # #   #  #  ### ##  ##   #   #   ## # # ##  #   # # ### ### ### " +
	"# #   #  #    #  #    # # #  #  # #   # ### # # # # # # #   #   ### " +
	" #    # ### ##   #  ##   #   #   #  ##  # # ##   #  ##  ### #   ### " +
	"                                                                    "

func hexfontGet(hex, x, y byte) bool {
	switch hex {
	case '0':
		hex = 0
	case '1':
		hex = 1
	case '2':
		hex = 2
	case '3':
		hex = 3
	case '4':
		hex = 4
	case '5':
		hex = 5
	case '6':
		hex = 6
	case '7':
		hex = 7
	case '8':
		hex = 8
	case '9':
		hex = 9
	case 'A', 'a':
		hex = 10
	case 'B', 'b':
		hex = 11
	case 'C', 'c':
		hex = 12
	case 'D', 'd':
		hex = 13
	case 'E', 'e':
		hex = 14
	case 'F', 'f':
		hex = 15
	default:
		hex = 16
	}
	if x >= 4 {
		return false
	}
	if y >= 6 {
		return false
	}

	return hexfont[4*int(hex)+int(y)*17*4+int(x)] == '#'
}

// GetRGBTexture returns the RGB texture for a given Unicode character.
// If the character is not in the font, it generates a placeholder using hex digits.
func (f *Font) GetRGBTexture(code string) [][3]byte {
	if f.mapping == nil {
		return nil
	}

	a, ok := f.mapping[code]
	if !ok {
		if f.cellx < 12 || f.celly < 24 {
			return nil
		}

		faketexture := make([][3]byte, f.cellx*f.celly)
		fakestring := fmt.Sprintf("%+q", code)

		fakestring = strings.Replace(fakestring, "\"", "", -1)
		fakestring = strings.Replace(fakestring, "\\", "", -1)
		fakestring = strings.Replace(fakestring, "u", "", -1)
		fakestring = strings.Replace(fakestring, "U", "", -1)

		var i = 0
		for xbox := byte(0); xbox < 3; xbox++ {
			for ybox := byte(0); ybox < 4; ybox++ {
				for y := byte(0); y < 6; y++ {
					for x := byte(0); x < 4; x++ {
						pos := int(ybox)*f.cellx*6 + int(xbox)*4 + int(y)*f.cellx + int(x)
						if len(fakestring) > i {
							if hexfontGet(fakestring[i], x, y) {
								faketexture[pos][0] = 255
								faketexture[pos][1] = 255
								faketexture[pos][2] = 255
							}
						}
					}
				}
				i++
			}
		}
		// memoization
		f.mapping[code] = faketexture
		return faketexture
	}
	return a
}

// CellWidth returns the width of a character cell in pixels.
func (f *Font) CellWidth() int {
	return f.cellx
}

// CellHeight returns the height of a character cell in pixels.
func (f *Font) CellHeight() int {
	return f.celly
}

// Load loads a font from an embedded PNG or JPEG file.
// descriptor is a tab/newline separated grid of characters matching the image layout.
// trailer is appended to each character code for aliasing.
func (f *Font) Load(name, descriptor, trailer string) error {
	file, err := embedFonts.Open("fonts/" + name)
	if err != nil {
		return fmt.Errorf("font not found: %s: %w", name, err)
	}
	defer file.Close()

	var img image.Image

	if strings.HasSuffix(name, ".png") {
		img, err = png.Decode(file)
		if err != nil {
			return fmt.Errorf("cannot decode png %s: %w", name, err)
		}
	} else {
		img, err = jpeg.Decode(file)
		if err != nil {
			return fmt.Errorf("cannot decode jpeg %s: %w", name, err)
		}
	}

	b := img.Bounds()

	var width = b.Max.X - b.Min.X
	var height = b.Max.Y - b.Min.Y

	var buffer = strings.Split(strings.ReplaceAll(descriptor, "\r\n", "\n"), "\n")
	var buf0 = strings.Split(buffer[0], "\t")

	var cellx = width / len(buf0)
	var celly = height / len(buffer)

	if f.mapping == nil {
		f.cellx = cellx
		f.celly = celly
	} else if f.cellx != cellx || f.celly != celly {
		return fmt.Errorf("only same cell sized fonts can be merged")
	}

	var mapping = make(map[string][2]int)
	var mapping2 = make(map[[2]int][][3]byte)

	for y, v := range buffer {
		var buf = strings.Split(strings.Trim(v, "\t"), "\t")
		for x, cell := range buf {
			mapping[cell] = [2]int{x, y}
		}
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		var iy = (y - b.Min.Y) / f.celly
		for x := b.Min.X; x < b.Max.X; x++ {
			var ix = (x - b.Min.X) / f.cellx
			var i = [2]int{ix, iy}

			var sli = mapping2[i]

			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)

			sli = append(sli, [3]byte{c.R, c.G, c.B})

			mapping2[i] = sli
		}
	}
	if f.mapping == nil {
		f.mapping = make(map[string][][3]byte)
	}
	for k, v := range mapping {
		f.mapping[k+trailer] = mapping2[v]
	}
	return nil
}

// NewFont creates a new Font and loads the basic ASCII font.
func NewFont() (*Font, error) {
	f := &Font{}
	err := f.Load("ascii.png", asciiDescriptor, "")
	if err != nil {
		return nil, err
	}
	return f, nil
}

func maxByte3(a, b [3]byte) [3]byte {
	return [3]byte{maxByte(a[0], b[0]), maxByte(a[1], b[1]), maxByte(a[2], b[2])}
}
func maxByte(a, b byte) byte {
	if a > b {
		return a
	}
	return b
}

func Each(descriptor string, function func(string) error) error {
	var buffer = strings.Split(strings.ReplaceAll(strings.ReplaceAll(descriptor, "\r\n", "\n"), "\t", "\n"), "\n")
	for _, v := range buffer {
		if len(v) == 0 {
			continue
		}
		err := function(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Font) Multiply(descriptor1, suffix, separator, descriptor2 string) error {
	err := Each(descriptor1, func(v string) error {
		err := f.Combine(suffix+v, descriptor2, v+separator)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (f *Font) Combine(combiner, descriptor, textureName string) error {
	if len(combiner) == 0 {
		println("no combiner, would create same named textures")
		return fmt.Errorf("no combiner")
	}

	var combinerTexture [][3]byte

	if len(textureName) == 0 {
		combinerTexture = f.GetRGBTexture(combiner)
	} else {
		combinerTexture = f.GetRGBTexture(textureName)
	}
	if len(combinerTexture) == 0 {
		println("no combiner texture")
		return fmt.Errorf("no combiner texture")
	}

	var buffer = strings.Split(strings.ReplaceAll(descriptor, "\r\n", "\n"), "\n")

	for _, v := range buffer {
		var buf = strings.Split(strings.Trim(v, "\t"), "\t")
		for _, cell := range buf {
			if len(cell) == 0 {
				continue
			}
			if cell == combiner {
				continue
			}

			var otherTexture = f.GetRGBTexture(cell)

			if len(otherTexture) != len(combinerTexture) {
				continue
			}

			var newTexture = make([][3]byte, len(otherTexture))
			for i := range newTexture {
				newTexture[i] = maxByte3(otherTexture[i], combinerTexture[i])
			}

			f.mapping[cell+combiner] = newTexture

			//println(cell+combiner)
		}
	}
	return nil
}


func (f *Font) Alias(alias, key string) error {
	if f.mapping == nil {
		println("no mapping")
		return fmt.Errorf("no mapping")
	}
	if f.mapping[key] == nil {
		println("key missing")
		return fmt.Errorf("key missing")
	}
	f.mapping[alias] = f.mapping[key]
	return nil
}

// LoadExtendedFonts loads additional Unicode font files.
// This is optional and can be called after NewFont() to support more characters.
// Note: Extended fonts must have the same cell dimensions as the base font.
func (f *Font) LoadExtendedFonts() error {
	_ = f.Load("ascii.png", asciiDescriptor, "")
	_ = f.Load("extendeda.png", extendedaDescriptor, "")
	_ = f.Load("extendedb.png", extendedbDescriptor, "")
	_ = f.Load("supplement.png", supplementDescriptor, "")
	_ = f.Load("spacingmod.png", spacingmodDescriptor, "")
	_ = f.Load("ipa.png", ipaDescriptor, "")
	_ = f.Load("greek.png", greekDescriptor, "")
	_ = f.Load("cyrillic.png", cyrillicDescriptor, "")
	_ = f.Load("vietnamese.png", vietnameseDescriptor, "")
	_ = f.Load("hangul0.png", hangul0Descriptor, "")
	_ = f.Load("hangul1.png", hangul0Descriptor, "1")
	_ = f.Load("hangul9.png", hangul9Descriptor, "")
	_ = f.Multiply(hangul0Descriptor, "x", "1", hangul9Descriptor)
	_ = Each(hangul0Descriptor, func(v string) error {
		const buf = "	\u11a8\u11a9\u11aa\u11ab\u11ac\u11ad\u11ae\u11af\u11b0\u11b1\u11b2" +
			"\u11b3\u11b4\u11b5\u11b6\u11b7\u11b8\u11b9\u11ba\u11bb\u11bc" +
			"\u11bd\u11be\u11bf\u11c0\u11c1\u11c2"
		for i := 1; i < 28; i++ {

			var target = string([]rune(v)[0] + rune(i))
			var bottom = string([]rune(buf)[i])

			//println(v, "|",  bottom + "x" + v)
			_ = f.Alias(target, bottom+"x"+v)
		}

		return nil
	})
	err := f.Load("combining.png", combiningDescriptor, "")
	if err != nil {
		println(err.Error())
	}
	_ = f.Multiply(combiningDescriptor, "", "", cyrillicDescriptor)
	_ = f.Load("armenian.png", armenianDescriptor, "")

	_ = f.Load("chinese1.jpg", chinese1Descriptor, "")

	_ = f.Load("devanagari1.png", devanagari1Descriptor, "")
	_ = f.Load("devanagari2.png", devanagari2Descriptor, "")
	_ = f.Load("devanagari3.png", devanagari3Descriptor, "")
	_ = f.Combine("ः", devanagari1Descriptor, "")
	_ = f.Combine("ं", devanagari1Descriptor, "")
	_ = f.Combine("ा", devanagari1Descriptor, "")
	_ = f.Combine("ऻ", devanagari1Descriptor, "")
	_ = f.Combine("ि", devanagari1Descriptor, "")
	_ = f.Combine("ी", devanagari1Descriptor, "")
	_ = f.Combine("े", devanagari1Descriptor, "")
	_ = f.Combine("ॅ", devanagari1Descriptor, "")
	_ = f.Combine("ॆ", devanagari1Descriptor, "")
	_ = f.Combine("ै", devanagari1Descriptor, "")
	_ = f.Combine("ॉ", devanagari1Descriptor, "")
	_ = f.Combine("ॊ", devanagari1Descriptor, "")
	_ = f.Combine("ो", devanagari1Descriptor, "")
	_ = f.Combine("ौ", devanagari1Descriptor, "")
	_ = f.Combine("़", devanagari1Descriptor, "")
	_ = f.Combine("ॎ", devanagari1Descriptor, "")
	_ = f.Combine("ऀ", devanagari1Descriptor, "")
	_ = f.Combine("ँ", devanagari1Descriptor, "")
	_ = f.Combine("ऺ", devanagari1Descriptor, "")
	_ = f.Combine("ु", devanagari1Descriptor, "")
	_ = f.Combine("ू", devanagari1Descriptor, "")
	_ = f.Combine("ृ", devanagari1Descriptor, "")
	_ = f.Combine("ॄ", devanagari1Descriptor, "")
	_ = f.Combine("ॏ", devanagari1Descriptor, "")
	_ = f.Combine("ॕ", devanagari1Descriptor, "")
	_ = f.Combine("ॖ", devanagari1Descriptor, "")
	_ = f.Combine("ॗ", devanagari1Descriptor, "")
	_ = f.Combine("ॢ", devanagari1Descriptor, "")
	_ = f.Combine("ॣ", devanagari1Descriptor, "")

	_ = f.Combine("ों", devanagari1Descriptor, "")
	_ = f.Combine("ें", devanagari1Descriptor, "")
	_ = f.Combine("़ा", devanagari1Descriptor, "")
	_ = f.Combine("ो़", devanagari1Descriptor, "")
	_ = f.Combine("़ि", devanagari1Descriptor, "")
	_ = f.Combine("ूँ", devanagari1Descriptor, "")
	_ = f.Combine("़ो", devanagari1Descriptor, "")

	_ = f.Combine("ꣿ", "ए", "")
	_ = f.Alias("ꣾ", "एꣿ")
	_ = f.Alias("क़्", "क़्")
	_ = f.Alias("ख़्", "ख़्")
	_ = f.Alias("ग़्", "ग़्")
	_ = f.Alias("ज़्", "ज़्")
	_ = f.Alias("ड़्", "ड़्")
	_ = f.Alias("ढ़्", "ढ़्")
	_ = f.Alias("फ़्", "फ़्")
	_ = f.Alias("य़्", "य़्")
	_ = f.Alias("ड़", "ड़")
	_ = f.Alias("ढ़", "ढ़")
	_ = f.Alias("ॴ", "आऺ")
	_ = f.Alias("ॶ", "अॖ")
	_ = f.Alias("ॷ", "अॗ")
	_ = f.Alias("ॵ", "अॏ")
	_ = f.Alias("ॲ", "अॅ")
	_ = f.Alias("ꣲ", "ँ")
	_ = f.Alias("॰", "°")

	_ = f.Alias("\t", " ")
	_ = f.Alias("", " ")
	return nil
}

