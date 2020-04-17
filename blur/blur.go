package blur

import (
	"math"
	"image"
	// "image/png"
	// "image/jpeg"
	// "os"
	// "flag"
)


type Skernel struct {
	X []uint32
	Y []uint32
}

const MAX int = 0xffffffff

var (
	DefaultStdDev float64 = 2
	Defaultk *Skernel
	DefaultR = image.Rect(0,0,MAX,MAX)
)

func init () {
	Defaultk = NewKernel(DefaultStdDev)
	
}

// func main() {
// 	fs,err := os.Open(*imgFile)
// 	defer fs.Close()
// 	if err!= nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// 	srcImg,_  :=  png.Decode(fs)
// 	dstImg := image.NewRGBA(srcImg.Bounds())
// 	GaussianBlur(dstImg, srcImg, k)
// 	fs1,err1 := os.OpenFile(*outFile, os.O_WRONLY|os.O_CREATE, 0666)
// 	defer fs1.Close()
// 	if err1!=nil {
// 		fmt.Println(err1)
// 		os.Exit(1)
// 	}
	
// 	png.Encode(fs1, dstImg)
// }




func NewKernel(sigma float64) *Skernel{
	sigmaMulSqrt2pi := 1 / (math.Sqrt(2 * math.Pi) * sigma)
	divSigmaPow2 := 1 / (2 * sigma * sigma)
	kerLen := int(math.Ceil(sigma * 6))
	kernel := make([]float64, 2* kerLen+1)
	var ikernel []uint32
	var sum float64 = 0
	for i:=0; i < kerLen; i++ {
		g := math.Exp(-float64(i * i) * divSigmaPow2) * sigmaMulSqrt2pi
		kernel[kerLen - i] = g
		kernel[kerLen + i] = g
		sum += g
		if i != 0 {
			sum += g
		}

	}
	sum = 1/sum
	for i,v := range kernel {
		kernel[i] = v * sum
		ik := uint32(kernel[i] * (1 << 8))
		if ik > 0 {
			ikernel = append(ikernel, ik)
		}
	}

	k := new(Skernel) 
	k.X = ikernel
	k.Y = ikernel

	return k
}


// 高斯模糊
// src	目标图片
// dst	操作的目标区域
// +--------------+
// |	+---+ src |
// |	|dst|     |
// |	+---+	  |
// +--------------+
func GaussianBlur(dst *image.RGBA, src image.Image, k *Skernel) {
	dstRect := dst.Bounds()

	width, height := dstRect.Dx(), dstRect.Dy()
	radius := (len(k.X)-1)/2

	bufTmp := make([]uint32, width*height*4)

	// vertical
	for y := dstRect.Min.Y; y < dstRect.Max.Y; y++ {
		for x := dstRect.Min.X; x < dstRect.Max.X; x++ {
			var r, g, b, a uint32
			
			k0 := k.Y[radius]
			for i := 1; i <= radius; i++ {
				// above
				f := k.Y[radius - i]
				if y - i < dstRect.Min.Y {
					k0 += f
				} else {
					or,og,ob,oa := src.At(x, y-i).RGBA()
					r += (or>>8) * f
					g += (og>>8) * f
					b += (ob>>8) * f
					a += (oa>>8) * f
				}
				
				// below
				f = k.Y[radius + i]
				if y + i > dstRect.Max.Y {
					k0 += f
				} else {
					or,og,ob,oa := src.At(x, y+i).RGBA()
					r += (or>>8) * f
					g += (og>>8) * f
					b += (ob>>8) * f
					a += (oa>>8) * f
				}
			}

			or,og,ob,oa := src.At(x, y).RGBA()
			r += (or>>8) * k0
			g += (og>>8) * k0
			b += (ob>>8) * k0
			a += (oa>>8) * k0

			ix := (y-dstRect.Min.Y)*width*4 + (x-dstRect.Min.X) * 4
			bufTmp[ix] = r>>8
			bufTmp[ix+1] = g>>8
			bufTmp[ix+2] = b>>8
			bufTmp[ix+3] = a>>8
		}
	}

	// horizontal
	for y := 0; y< height;y++ {
		for x:=0; x< width; x++ {
			var r, g, b, a uint32
			k0 := k.X[radius]
			off := (y*width + x) * 4
			for i:= 1; i<=radius; i++ {
				// left
				f := k.X[radius - i]
				if x-i < 0 {
					k0 += f
				}else {
					ix := off - i * 4
					r += bufTmp[ix] * f
					g += bufTmp[ix+1] * f
					b += bufTmp[ix+2] * f
					a += bufTmp[ix+3] * f
				}

				// right
				f = k.X[radius + i]
				if x+i >= width {
					k0 += f
				} else {
					ix := off + i * 4
					r += bufTmp[ix] * f
					g += bufTmp[ix+1] * f
					b += bufTmp[ix+2] * f
					a += bufTmp[ix+3] * f
				}
			}
			r += bufTmp[off] * k0
			g += bufTmp[off+1] * k0
			b += bufTmp[off+2] * k0
			a += bufTmp[off+3] * k0

			dstoff := y*width*4 + x*4
			dst.Pix[dstoff] = uint8(clamp(float64(r>>8), 0, 0xff))
			dst.Pix[dstoff+1] = uint8(clamp(float64(g>>8), 0, 0xff))
			dst.Pix[dstoff+2] = uint8(clamp(float64(b>>8), 0, 0xff))
			dst.Pix[dstoff+3] = uint8(clamp(float64(a>>8), 0, 0xff))
		}
	}
}


func clamp (a,b,c float64) float64 {
	if a < b {
		return b
	}
	if a > c {
		return c
	}
	return a
}

func GaussianBlurImage(src image.Image, r image.Rectangle, std float64) image.Image {
	srcBound := src.Bounds()
	x0 := r.Min.X 
	y0 := r.Min.Y
	x1 := r.Max.X
	y1 := r.Max.Y
	if  r.Min.X < srcBound.Min.X {
		x0 = srcBound.Min.X
	}
	if r.Min.Y < srcBound.Min.Y {
		y0 = srcBound.Min.Y
	}
	if r.Max.X > srcBound.Max.X {
		x1 = srcBound.Max.X
	}
	if r.Max.Y > srcBound.Max.Y {
		y1 = srcBound.Max.Y
	}

	dst := image.NewRGBA(image.Rect(x0, y0, x1, y1))
	k := NewKernel(std)
	GaussianBlur(dst, src, k)
	return dst
}