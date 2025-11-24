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

//go:embed fonts/*.png
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

// LoadExtendedFonts loads additional Unicode font files.
// This is optional and can be called after NewFont() to support more characters.
// Note: Extended fonts must have the same cell dimensions as the base font.
func (f *Font) LoadExtendedFonts() error {
	// Try to load supplement (same cell size as ASCII)
	if err := f.Load("supplement.png", supplementDescriptor, ""); err != nil {
		// Extended fonts are optional, errors are not fatal
		return fmt.Errorf("failed to load supplement.png: %w", err)
	}
	
	// Extended fonts have different cell sizes and can't be merged
	// They would need to be loaded into a separate Font instance
	
	return nil
}

// asciiDescriptor defines the layout of the ASCII font image
const asciiDescriptor = ` 	!	"	#	$	%	&	'	(	)	*	+	,	-	.	/
0	1	2	3	4	5	6	7	8	9	:	;	<	=	>	?
@	A	B	C	D	E	F	G	H	I	J	K	L	M	N	O
P	Q	R	S	T	U	V	W	X	Y	Z	[	\	]	^	_
` + "`" + `	a	b	c	d	e	f	g	h	i	j	k	l	m	n	o
p	q	r	s	t	u	v	w	x	y	z	{	|	}	~	`

// extendedaDescriptor defines extended Latin characters
const extendedaDescriptor = "" +
	`Ā	Ġ	Ŀ	ş	ā	ġ	ŀ	Š
	Ă	Ģ	Ł	š	ă	ģ	ł	Ţ
	Ą	Ĥ	Ń	ţ	ą	ĥ	ń	Ť
	Ć	Ħ	Ņ	ť	ć	ħ	ņ	Ŧ
	Ĉ	Ĩ	Ň	ŧ	ĉ	ĩ	ň	Ũ
	Ċ	Ī	ŉ	ũ	ċ	ī	Ŋ	Ū
	Č	Ĭ	ŋ	ū	č	ĭ	Ō	Ŭ
	Ď	Į	ō	ŭ	ď	į	Ŏ	Ů
	Đ	İ	ŏ	ů	đ	ı	Ő	Ű
	Ē	Ĳ	ő	ű	ē	ĳ	Œ	Ų
	Ĕ	Ĵ	œ	ų	ĕ	ĵ	Ŕ	Ŵ
	Ė	Ķ	ŕ	ŵ	ė	ķ	Ŗ	Ŷ
	Ę	ſ	ŗ	ŷ	ę	ĸ	Ř	Ÿ
	Ě	Ĺ	ř	Ź	ě	ĺ	Ś	ź
	Ĝ	Ļ	ś	Ż	ĝ	ļ	Ŝ	ż
	Ğ	Ľ	ŝ	Ž	ğ	ľ	Ş	ž`

// supplementDescriptor defines supplemental Latin characters
const supplementDescriptor = "" +
	`¡	±	À	Ï	à	ï
	¢	²	Á	Ñ	á	ñ
	£	³	Â	Ò	â	ò
	¤	´	Ã	Ó	ã	ó
	¥	µ	Ä	Ô	ä	ô
	¦	¶	Å	Õ	å	õ
	§	·	Æ	Ö	æ	ö
	¨	¸	Ç	Ø	ç	ø
	©	¹	Ð	Ù	ð	ù
	ª	º	È	Ú	è	ú
	«	»	É	Û	é	û
	¬	¼	Ê	Ü	ê	ü
	½	¾	Ë	Ý	ë	ý
	­®	¿	Ì	Ÿ	ì	ÿ
	¯	÷	Í	Þ	í	þ
	°	×	Î	ẞ	î	ß`

// extendedbDescriptor defines more extended Latin characters
const extendedbDescriptor = "" +
	`ǎ	Ǎ	ǳ	ǲ	ȟ	Ȟ	ǒ	Ǒ	ȑ	Ȑ	ȳ	Ȳ
	ǻ	Ǻ	Ǳ	ǆ	ƕ	Ƕ	ȫ	Ȫ	ȓ	Ȓ	ɏ	Ɏ
	ǟ	Ǟ	ǅ	Ǆ	ǐ	Ǐ	ȭ	Ȭ	ɍ	Ɍ	ƴ	Ƴ
	ȧ	Ȧ	Ɖ	Ɗ	ȉ	Ȉ	ȯ	Ȯ	ș	Ș	ȝ	Ȝ
	ǡ	Ǡ	ȩ	Ȩ	ȋ	Ȋ	ȱ	Ȱ	Ʀ	ȿ	ƶ	Ƶ
	ȁ	Ȁ	ȅ	Ȅ	Ɨ	Ɩ	ǿ	Ǿ	ț	Ț	ȥ	Ȥ
	ȃ	Ȃ	ȇ	Ȇ	ǰ	ȷ	ǫ	Ǫ	ƾ	Ⱦ	ǯ	Ǯ
	ǽ	Ǽ	ɇ	Ɇ	ɉ	Ɉ	ǭ	Ǭ	Ƭ	Ʈ	ƹ	Ƹ
	ǣ	Ǣ	ǝ	Ǝ	ǩ	Ǩ	ȍ	Ȍ	ǔ	Ǔ	ƿ	Ƿ
	Ⱥ	ƀ	Ə	Ɛ	ƙ	Ƙ	ȏ	Ȏ	ǘ	Ǘ	ǜ	Ǜ
	Ƀ	Ɓ	ƒ	Ƒ	ǉ	ǈ	Ɲ	ȵ	ǚ	Ǚ	ƽ	Ƽ
	ƃ	Ƃ	ǵ	Ǵ	ƚ	Ƚ	Ǉ	Ɔ	ǖ	Ǖ	ƅ	Ƅ
	ȼ	Ȼ	ǧ	Ǧ	ȴ	ƛ	ơ	Ơ	ȕ	Ȕ	ƨ	Ƨ
	ƈ	Ƈ	ǥ	Ǥ	ǹ	Ǹ	ȣ	Ȣ	ȗ	Ȗ	ȡ	ƫ
	ȸ	ȹ	Ɠ	Ɣ	ǌ	ǋ	ƥ	Ƥ	ɋ	Ɋ	ư	Ư
	ƌ	Ƌ	ƣ	Ƣ	ƞ	Ƞ	Ǌ	Ɵ	Ʃ	ƪ	ƭ	ȶ`
