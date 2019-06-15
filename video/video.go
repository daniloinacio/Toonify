package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"gocv.io/x/gocv"
)

func main() {
	// Get initial time
	start := time.Now()

	// Open video capture
	video, _ := gocv.OpenVideoCapture("xerek720.mp4")
	// Define destiny FPS
	DestFPS := 15
	// Get source FPS
	SrcFPS := video.Get(gocv.VideoCaptureFPS)
	// Get total amount of frames
	FramesTot := video.Get(gocv.VideoCaptureFrameCount)
	// Get video lenght
	video.Set(gocv.VideoCapturePosAVIRatio, 1)
	videoLen := video.Get(gocv.VideoCapturePosMsec)
	video.Set(gocv.VideoCapturePosAVIRatio, 0)
	// Get frame lenght
	frameLen := videoLen / FramesTot
	// Get frame counter
	SrcFPS = math.Round(SrcFPS)
	FrameCont := SrcFPS / float64(DestFPS)
	FrameCont = FrameCont * frameLen

	// Initialize output
	vidWidth := int(video.Get(gocv.VideoCaptureFrameWidth))
	vidHeight := int(video.Get(gocv.VideoCaptureFrameHeight))
	output, _ := gocv.VideoWriterFile("xerek-15.avi", "MJPG", float64(DestFPS), vidWidth, vidHeight, true)

	window := gocv.NewWindow("Hello")
	window2 := gocv.NewWindow("Hello2")
	img := gocv.NewMat()
	ret := true
	FrameIt := 0.0
	frameS := video.Get(gocv.VideoCapturePosMsec)

	for {
		frameS = video.Get(gocv.VideoCapturePosMsec)
		fmt.Printf("time: %v / %v\n", frameS, videoLen)
		ret = video.Read(&img)
		if !ret {
			break
		}
		img2 := toonify(img)
		window.IMShow(img)
		window2.IMShow(img2)
		output.Write(img2)
		img2.Close()
		FrameIt = FrameIt + FrameCont
		video.Set(gocv.VideoCapturePosMsec, FrameIt)
		window.WaitKey(1000 / DestFPS)
	}

	video.Close()

	time := time.Since(start)
	log.Printf("Progam took %s\n", time)
}
