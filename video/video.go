package main

import (
	"gocv.io/x/gocv"
)

func main() {
	//webcam, _ := gocv.OpenVideoCapture("[AnimesTelecine] Isekai Quartet - 05 [720p].mp4")
	webcam, _ := gocv.OpenVideoCapture(0)
	window := gocv.NewWindow("Hello")
	img := gocv.NewMat()
	fps := 60

	for {
		webcam.Read(&img)
		window.IMShow(img)
		window.WaitKey(1000 / fps)
	}
}
