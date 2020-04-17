package vertifyCodeImg

import (
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/font"
	"image"
	"image/draw"
	"image/color"
	"vertifyCodeImg/blur"
	"math/rand"
	"time"
	// "fmt"
)


var (
	W int = 20
	H int = 80
	Bg = image.NewUniform(color.RGBA{205, 205, 205, 255})
	Edge int = 4
	Dpi float64 = 70
	WordPool []string
	ColorRange = []int{45, 155}	// 47~200
	SizeRange = []float64{40, 30}	// 40~70
	BlurRange = []float64{0.5, 1}  	//0.5+1.5
	RotateRange int = 60
	ScaleRange = []float64{0.8, 0.5} // 0.8~1.3
	WordNum int = 5

	NoiseVar float64 = 25
)

var FI *FontImage

func Init(){
	FI = NewFontImage(goregular.TTF, Dpi, Edge)
	for i:=49;i<=57;i++ {
		WordPool = append(WordPool, string(i))
	}
	for i:=65;i<=90;i++ {
		WordPool = append(WordPool, string(i))
	}
	for i:=97;i<=122;i++ {
		WordPool = append(WordPool, string(i))
	}
}


func CreateVertifyCode () (image.Image, []string){
	rand.Seed(time.Now().Unix())
	wordImages := make([]*Word, WordNum)
	words := make([]string, WordNum)
	for i:=0;i<WordNum;i++ {
		wordImages[i] = RandomWord()
		words[i] = wordImages[i].W
	}

	W += 20
	dc := gg.NewContext(W, H)
	dc.DrawImage(Bg, 0, 0)
	dc.Stroke()
	// draw word
	for i:=0; i<WordNum; i++ {
		dc.RotateAbout(gg.Radians(wordImages[i].Rotate), wordImages[i].Rx, wordImages[i].Ry)
		// outline
		// dc.SetHexColor("#ff0000")
		// dc.SetLineWidth(1)
		// dc.DrawRectangle(float64(wordImages[i].Ix), float64(wordImages[i].Iy), float64(FI.Width), float64(FI.Height))
		// dc.Stroke()
		// end
		dc.DrawImage(wordImages[i].Image, wordImages[i].Ix, wordImages[i].Iy)
		dc.RotateAbout(gg.Radians(-wordImages[i].Rotate), wordImages[i].Rx, wordImages[i].Ry)
	}

	// draw line
	for i:=0; i<WordNum; i++ {
		dc.RotateAbout(gg.Radians(wordImages[i].Rotate), wordImages[i].Rx, wordImages[i].Ry)
		// outline
		// dc.SetHexColor("#ff0000")
		// dc.SetLineWidth(1)
		// dc.DrawRectangle(float64(wordImages[i].Ix), float64(wordImages[i].Iy), float64(FI.Width), float64(FI.Height))
		// dc.Stroke()
		// end
		dc.DrawImage(wordImages[i].Image, wordImages[i].Ix, wordImages[i].Iy)
		dc.RotateAbout(gg.Radians(-wordImages[i].Rotate), wordImages[i].Rx, wordImages[i].Ry)
	}

	PointNoise(dc.Image().(*image.RGBA))
	LineNoise(dc)

	return dc.Image(), words
}


type Word struct {
	W string
	Image	*image.RGBA
	Ix, Iy		int
	Rotate	float64
	Rx, Ry	float64
	ScaleX, ScaleY float64

}


func RandomWord () *Word {
	word := &Word{}
	word.W = WordPool[rand.Intn(len(WordPool))]
	size := SizeRange[0] + SizeRange[1] * rand.Float64()
	r := uint8(ColorRange[0] + rand.Intn(ColorRange[1]))
	g := uint8(ColorRange[0] + rand.Intn(ColorRange[1]))
	b := uint8(ColorRange[0] + rand.Intn(ColorRange[1]))
	// a := uint8(ColorRange[0] + rand.Intn(ColorRange[1]))
	fg := image.NewUniform(color.RGBA{r,g,b,255})
	img := FI.CreateImage(word.W, size, fg, nil)
	std := BlurRange[0] + rand.Float64() * BlurRange[1]
	word.Image = blur.GaussianBlurImage(img, blur.DefaultR, std).(*image.RGBA)
	// word.Ix += int(FI.Width + rand.Intn(FI.Width/3) - FI.Width/2 )
	word.Ix = W
	W += FI.Width
	word.Iy = rand.Intn((H - FI.Height)/2)
	
	word.Rotate = float64(RotateRange - rand.Intn(RotateRange*2))
	word.Rx = float64(word.Ix + FI.Width/2)
	word.Ry = float64(word.Iy + FI.Height/2)
	word.ScaleX = ScaleRange[0] + rand.Float64() * ScaleRange[1]
	word.ScaleY = ScaleRange[0] + rand.Float64() * ScaleRange[1]
	// fmt.Println(w, word.Ix, word.Iy)
	return word
}

// 噪声
func PointNoise(img *image.RGBA) {
	k := func(x uint32) uint8{
		return uint8(rand.NormFloat64()*NoiseVar + float64(x))
	}

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			or,og,ob,oa := img.At(x,y).RGBA()
			r := k(or>>8)
			g := k(og>>8)
			b := k(ob>>8)
			a := uint8(oa>>8)
			img.Set(x, y, color.RGBA{r,g,b,a})
		}
	}
}

// 线条
func LineNoise(dc *gg.Context) {
	for i:=0;i<10;i++ {
		x1 := rand.Float64() * float64(W)
		y1 := rand.Float64() * float64(H)
		x2 := rand.Float64() * float64(W)
		y2 := rand.Float64() * float64(H)
		r := rand.Float64()
		g := rand.Float64()
		b := rand.Float64()
		a := rand.Float64()*0.3 + 0.3
		w := rand.Float64()*4 + 1
		dc.SetRGBA(r, g, b, a)
		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}
	
}

// 侵蚀
func ones() {

}




type FontImage struct {
	FontTTF			*truetype.Font
	Dpi				float64
	Width, Height		int
	Edge				int
	Fg, Bg			image.Image
}

func NewFontImage (tff []byte, dpi float64, edge int) *FontImage{
	f := new(FontImage)
	f.FontTTF, _ = truetype.Parse(tff)
	f.Edge = edge
	f.Dpi = dpi
	f.Fg = image.Black
	f.Bg = image.Transparent 
	return f
}

func (f *FontImage) ComPos(str string, size float64) (int, int){
	face := truetype.NewFace(f.FontTTF, &truetype.Options{
		Size:    size,
		DPI:     f.Dpi,
	})
	d := &font.Drawer{
		Face: face,
	}
	dBound,enhance := d.BoundString(str)
	f.Width = int(enhance>>6) + f.Edge*2
	f.Height = int(dBound.Max.Y >> 6)-int(dBound.Min.Y >> 6) +f.Edge*2
	// fmt.Println(int(dBound.Min.X >> 6), int(dBound.Min.Y >> 6) )
	// fmt.Println(int(dBound.Max.X>>6), int(dBound.Max.Y>>6 ))
	// fmt.Println(int(enhance>>6))
	// fmt.Println(f.Width, f.Height)

	return f.Edge, f.Edge-int(dBound.Min.Y >> 6)
}

func (f *FontImage) CreateImage(str string, size float64, fg image.Image, bg image.Image) image.Image {
	if fg == nil {
		fg = f.Fg
	}
	if bg == nil {
		bg = f.Bg
	}
	x0, y0 := f.ComPos(str, size)
	rgba := image.NewRGBA(image.Rect(0, 0, f.Width, f.Height))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	d := &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(f.FontTTF, &truetype.Options{
			Size:    size,
			DPI:     f.Dpi,
		}),
	}

	d.Dot = fixed.P(x0, y0)
	d.DrawString(str)
	return rgba
}
