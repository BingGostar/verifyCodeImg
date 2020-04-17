生成图片验证码

```
import (
	"image/png"
	"vertifyCodeImg"
	"os"
	"fmt"
)

// 初始化
vertifyCodeImg.Init()

// 生成图片和验证码
vertifyImg, words := vertifyCodeImg.CreateVertifyCode()

fp, _ := os.Create("test.png")
defer fp.Close()
png.Encode(fp, vertifyImg)
fmt.Println(words)
```