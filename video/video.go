package main

import (
	"gocv.io/x/gocv"
)

func main() {
	fps := 15
	// Open video capture
	webcam, _ := gocv.OpenVideoCapture("keyboard.mp4")
	// Initialize output
	vid_width := int(webcam.Get(3))
	vid_height := int(webcam.Get(4))
	output, _ := gocv.VideoWriterFile("ToonVideo.avi","MJPG",float64(fps),vid_width,vid_height,true)

	window := gocv.NewWindow("Hello")
	window2 := gocv.NewWindow("Hello2")
	img := gocv.NewMat()
	ret := true

	for {
		ret = webcam.Read(&img)
		if !ret{
			break;
		}
		img2 := toonify(img)
		window.IMShow(img)
		window2.IMShow(img2)
		output.Write(img2)
		window.WaitKey(1000 / fps)
	}

	webcam.Close()
	output.Close()

}
