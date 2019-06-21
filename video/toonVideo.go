package main

import (
	"Toonify/kmeans"
	"fmt"
	"log"
	"math"
	"time"

	"gocv.io/x/gocv"
)

//var wg2 sync.WaitGroup
//var wg3 sync.WaitGroup
//var wg4 sync.WaitGroup
//var wgMain sync.WaitGroup

// Formata a imagem para ser processada pelo kmeans
func formatData(img gocv.Mat) []kmeans.ClusteredObservation {
	imgFloat64 := gocv.NewMat()
	defer imgFloat64.Close()
	img.ConvertTo(&imgFloat64, gocv.MatTypeCV64F)
	slice, _ := imgFloat64.DataPtrFloat64()

	rows := img.Rows() * img.Cols()
	cols := img.Channels()

	data := make([]kmeans.ClusteredObservation, rows)
	for i := 0; i < rows; i++ {
		aux := make([]float64, cols)
		for j := 0; j < cols; j++ {
			aux[j] = slice[i*cols+j]
		}
		data[i].Observation = aux
	}
	return data
}

//Reconstroi a imagem a partir da clusterização do kmeans
func imgRework(clusteredData []kmeans.ClusteredObservation, centroids []kmeans.Observation, rows int, cols int) (gocv.Mat, error) {
	//cria uma slice de Mats de 1 canal
	mat := make([]gocv.Mat, 3)
	defer mat[0].Close()
	defer mat[1].Close()
	defer mat[2].Close()
	mat[0] = gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
	mat[1] = gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
	mat[2] = gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			index := clusteredData[i*cols+j].ClusterNumber
			//valores do pixel na posição i, j nos 3 canais
			ch := centroids[index]
			//seta os valores nos canais correspondentes
			mat[0].SetUCharAt(i, j, uint8(ch[0]))
			mat[1].SetUCharAt(i, j, uint8(ch[1]))
			mat[2].SetUCharAt(i, j, uint8(ch[2]))
		}
	}
	img := gocv.NewMat()
	gocv.Merge(mat, &img)
	return img, nil
}

func stage1(video *gocv.VideoCapture) chan gocv.Mat {
	//frameS := video.Get(gocv.VideoCapturePosMsec)

	out := make(chan gocv.Mat, 1)
	go func() {
		cont := 0
		for {
			img := gocv.NewMat()
			imgBlured := gocv.NewMat()
			ret := video.Read(&img)
			if !ret {
				break
			}
			gocv.MedianBlur(img, &imgBlured, 7)
			img.Close()
			//wg2.Wait()
			cont++
			fmt.Printf("stage 1 %d\n", cont)
			out <- imgBlured
		}
		close(out)
	}()
	return out
}

func stage2(imgChan chan gocv.Mat) (chan []kmeans.ClusteredObservation, chan gocv.Mat) {
	out1 := make(chan []kmeans.ClusteredObservation, 1)
	out2 := make(chan gocv.Mat, 1)
	go func() {
		cont := 0
		for {
			imgEdges := gocv.NewMat()
			imgFiltered := gocv.NewMat()
			imgBlured, ok := <-imgChan
			if !ok {
				break
			}
			//wg2.Add(1)
			gocv.Canny(imgBlured, &imgEdges, 62.5, 125)
			gocv.BitwiseNot(imgEdges, &imgEdges)
			gocv.CvtColor(imgEdges, &imgEdges, gocv.ColorGrayToBGR)
			gocv.BilateralFilter(imgBlured, &imgFiltered, 7, 35, 35)
			imgBlured.Close()
			data := formatData(imgFiltered)
			imgFiltered.Close()
			//wg3.Wait()
			cont++
			fmt.Printf("stage 2 %d\n", cont)
			out2 <- imgEdges
			out1 <- data
			//wg2.Done()
		}
		close(out1)
		close(out2)
	}()
	return out1, out2
}

func stage3(dataChan chan []kmeans.ClusteredObservation, imgEdgesChan chan gocv.Mat) (chan []kmeans.ClusteredObservation, chan []kmeans.Observation, chan gocv.Mat) {
	out1 := make(chan []kmeans.ClusteredObservation, 1)
	out2 := make(chan []kmeans.Observation, 1)
	out3 := make(chan gocv.Mat, 1)
	go func() {
		cont := 0
		for {
			data, ok := <-dataChan
			if !ok {
				break
			}
			//wg3.Add(1)
			clusteredData, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
			//wg4.Wait()
			cont++
			fmt.Printf("stage 3 %d\n", cont)
			out1 <- clusteredData
			out2 <- centroids
			out3 <- <-imgEdgesChan
			//wg3.Done()
		}
		close(out1)
		close(out2)
		close(out3)
	}()
	return out1, out2, out3
}

func stage4(clusteredDataChan chan []kmeans.ClusteredObservation, centroidsChan chan []kmeans.Observation, imgChan chan gocv.Mat) chan gocv.Mat {
	out := make(chan gocv.Mat, 1)
	go func() {
		cont := 0
		for {
			imgToonify := gocv.NewMat()
			clusteredData, ok := <-clusteredDataChan
			imgEdges := <-imgChan
			centroids := <-centroidsChan
			if !ok {
				break
			}
			//wg4.Add(1)
			imgQuantized, _ := imgRework(clusteredData, centroids, imgEdges.Rows(), imgEdges.Cols())
			gocv.BitwiseAnd(imgEdges, imgQuantized, &imgToonify)
			imgQuantized.Close()
			imgEdges.Close()
			//wgMain.Wait()
			cont++
			fmt.Printf("stage 4 %d\n", cont)
			out <- imgToonify
			//wg4.Done()
		}
		close(out)
	}()
	return out
}

func main() {
	// Get initial time
	start := time.Now()

	// Open video capture
	video, _ := gocv.OpenVideoCapture("xerek360.mp4")
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
	FrameIt := 0.0

	// Pipeline
	chan1 := stage1(video)
	chan2, chan3 := stage2(chan1)
	chan4, chan5, chan6 := stage3(chan2, chan3)
	imgToonifyChan := stage4(chan4, chan5, chan6)

	for img := range imgToonifyChan {
		//wgMain.Add(1)

		window.IMShow(img)
		output.Write(img)
		FrameIt = FrameIt + FrameCont
		video.Set(gocv.VideoCapturePosMsec, FrameIt)
		window.WaitKey(1000 / DestFPS)
		img.Close()
		//wgMain.Done()
	}
	video.Close()

	time := time.Since(start)
	log.Printf("Progam took %s\n", time)
}
