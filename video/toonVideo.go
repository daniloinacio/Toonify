package main

import (
	"Toonify/kmeans"
	"log"
	"math"
	"time"

	"gocv.io/x/gocv"
)

// Nesse primeiro estágio é feita a captura de um frame do video
// e a aplicaão do filtro de média
func stage1(video *gocv.VideoCapture) chan gocv.Mat {

	out := make(chan gocv.Mat, 1)
	go func() { // goroutine
		for {
			img := gocv.NewMat()
			imgBlured := gocv.NewMat()
			ret := video.Read(&img)
			if !ret { // Encerra o loop quando chega ao fim do vídeo
				break
			}
			gocv.MedianBlur(img, &imgBlured, 7)
			img.Close()
			out <- imgBlured
		}
		close(out) // Encerra channel de saida
	}()
	return out
}

// Nesse estágio é extraido as bordas da frame e
// gerado o data a partir do frame para ser processado pelo kmeans
func stage2(imgChan chan gocv.Mat) (chan []kmeans.ClusteredPixel, chan gocv.Mat) {
	out1 := make(chan []kmeans.ClusteredPixel, 1) //Channels de saida
	out2 := make(chan gocv.Mat, 1)
	go func() { // goroutine
		for {
			imgEdges := gocv.NewMat()
			imgFiltered := gocv.NewMat()
			imgBlured, ok := <-imgChan
			if !ok { // Encerra o loop quando o channel estiver vazio e tiver sido fechado
				break
			}
			gocv.Canny(imgBlured, &imgEdges, 62.5, 125)
			gocv.BitwiseNot(imgEdges, &imgEdges)
			gocv.CvtColor(imgEdges, &imgEdges, gocv.ColorGrayToBGR)
			gocv.BilateralFilter(imgBlured, &imgFiltered, 7, 35, 35)
			imgBlured.Close()
			data := kmeans.FormatData(imgFiltered)
			imgFiltered.Close()
			out2 <- imgEdges
			out1 <- data

		}
		close(out1) // Encerra channels de saida
		close(out2)
	}()
	return out1, out2
}

// Nesse estágio é aplicado o kmeans no frame formatado
// no estágio anterior
func stage3(dataChan chan []kmeans.ClusteredPixel, imgEdgesChan chan gocv.Mat) (chan []kmeans.ClusteredPixel, chan []kmeans.Pixel, chan gocv.Mat) {
	out1 := make(chan []kmeans.ClusteredPixel, 1) //Channels de saida
	out2 := make(chan []kmeans.Pixel, 1)
	out3 := make(chan gocv.Mat, 1)
	go func() { // goroutine
		for {
			data, ok := <-dataChan
			if !ok { // Encerra o loop quando o channel estiver vazio e tiver sido fechado
				break
			}
			clusteredData, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
			out1 <- clusteredData
			out2 <- centroids
			out3 <- <-imgEdgesChan
		}
		close(out1) // Encerra channels de saida
		close(out2)
		close(out3)
	}()
	return out1, out2, out3
}

// Nesse último estágio o frame clusterizado é reconstruido
// e unido com as bordas
func stage4(clusteredDataChan chan []kmeans.ClusteredPixel, centroidsChan chan []kmeans.Pixel, imgChan chan gocv.Mat) chan gocv.Mat {
	out := make(chan gocv.Mat, 1) //Channel de saida
	go func() {                   // goroutine
		for {
			imgToonify := gocv.NewMat()
			clusteredData, ok := <-clusteredDataChan
			imgEdges := <-imgChan
			centroids := <-centroidsChan
			if !ok { // Encerra o loop quando o channel estiver vazio e tiver sido fechado
				break
			}
			imgQuantized, _ := kmeans.ImgRework(clusteredData, centroids, imgEdges.Rows(), imgEdges.Cols())
			gocv.BitwiseAnd(imgEdges, imgQuantized, &imgToonify)
			imgQuantized.Close()
			imgEdges.Close()
			out <- imgToonify

		}
		close(out) // Encerra channel de saida
	}()
	return out
}

func main() {

	start := time.Now()

	// Abre o video
	video, _ := gocv.OpenVideoCapture("xerek360.mp4")
	// FPS de destino
	DestFPS := 15
	// FPS original
	SrcFPS := video.Get(gocv.VideoCaptureFPS)
	// Total de frames de video
	FramesTot := video.Get(gocv.VideoCaptureFrameCount)
	// Duração do video
	video.Set(gocv.VideoCapturePosAVIRatio, 1)
	videoLen := video.Get(gocv.VideoCapturePosMsec)
	video.Set(gocv.VideoCapturePosAVIRatio, 0)
	// Duração do frame
	frameLen := videoLen / FramesTot
	// Contador de frames
	SrcFPS = math.Round(SrcFPS)
	FrameCont := SrcFPS / float64(DestFPS)
	FrameCont = FrameCont * frameLen

	// Inicializa a saida
	vidWidth := int(video.Get(gocv.VideoCaptureFrameWidth))
	vidHeight := int(video.Get(gocv.VideoCaptureFrameHeight))
	output, _ := gocv.VideoWriterFile("xerek-15.avi", "MJPG", float64(DestFPS), vidWidth, vidHeight, true)

	window := gocv.NewWindow("ToonVideo")
	FrameIt := 0.0

	// Pipeline
	chan1 := stage1(video)
	chan2, chan3 := stage2(chan1)
	chan4, chan5, chan6 := stage3(chan2, chan3)
	imgToonifyChan := stage4(chan4, chan5, chan6)

	// Exibe e salva os frames processados na pipeline
	for img := range imgToonifyChan {

		window.IMShow(img)
		output.Write(img)
		FrameIt = FrameIt + FrameCont
		video.Set(gocv.VideoCapturePosMsec, FrameIt)
		window.WaitKey(1000 / DestFPS)
		img.Close()
	}
	video.Close()

	time := time.Since(start)
	log.Printf("Progam took %s\n", time)
}
