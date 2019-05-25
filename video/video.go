package main

import (
	"gocv.io/x/gocv"
)

func main() {
	webcam, _ := gocv.OpenVideoCapture("keyboard.mp4")
	window := gocv.NewWindow("Hello")
	window2 := gocv.NewWindow("Hello2")
	img := gocv.NewMat()
	fps := 5

	for {
		webcam.Read(&img)
		img2 := toonify(img)
		window.IMShow(img)
		window2.IMShow(img2)
		window.WaitKey(1000 / fps)
	}

}
