package main

import (
	"Toonify/kmeans"
	"fmt"
	"log"
	"time"

	"gocv.io/x/gocv"
)

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

// Formata a imagem para ser processada pelo kmeans
func formatData(img gocv.Mat) []kmeans.ClusteredObservation {
	imgFloat64 := gocv.NewMat()
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

func makeEdges(imgBlured gocv.Mat, edgesChan chan gocv.Mat) {
	start := time.Now()
	imgEdges := gocv.NewMat()
	gocv.Canny(imgBlured, &imgEdges, 62.5, 125)
	elapsed := time.Since(start)
	log.Printf("canny time: %s", elapsed)
	gocv.BitwiseNot(imgEdges, &imgEdges)
	gocv.CvtColor(imgEdges, &imgEdges, gocv.ColorGrayToBGR)
	fmt.Println("Image edged")
	edgesChan <- imgEdges
}

func filter(imgBlured gocv.Mat, filterChan chan gocv.Mat) {
	imgFiltered := gocv.NewMat()
	start := time.Now()
	gocv.BilateralFilter(imgBlured, &imgFiltered, 7, 35, 35)
	elapsed := time.Since(start)
	log.Printf("bilateral time: %s", elapsed)
	fmt.Println("Image filtered")
	filterChan <- imgFiltered
}

func makeToon(img gocv.Mat, edgesChan chan gocv.Mat, filterChan chan gocv.Mat, toonChan chan gocv.Mat) {
	imgKmeans := gocv.NewMat()
	defer imgKmeans.Close()

	filteredImg := <-filterChan
	defer filteredImg.Close()
	start := time.Now()
	data := formatData(filteredImg)
	elapsed := time.Since(start)
	log.Printf("format time: %s", elapsed)
	start = time.Now()
	clusteredData, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
	elapsed = time.Since(start)
	log.Printf("kmeans time: %s", elapsed)
	start = time.Now()
	imgQuantized, _ := imgRework(clusteredData, centroids, img.Rows(), img.Cols())
	elapsed = time.Since(start)
	log.Printf("imgRework time: %s", elapsed)
	defer imgQuantized.Close()

	imgToonify := gocv.NewMat()
	gocv.BitwiseAnd(<-edgesChan, imgQuantized, &imgToonify)
	fmt.Println("Image Toonified")
	toonChan <- imgToonify
}

func main() {
	start := time.Now()
	// declarações de canais
	doneChan := make(chan string)     // para indicar que o programa terminou
	edgesChan := make(chan gocv.Mat)  // para colocar a imagem das bordas
	filterChan := make(chan gocv.Mat) // para a imagem filtrada
	toonChan := make(chan gocv.Mat)   // para a imagem cartunizada
	fmt.Println("Toonifying...")

	go func() {

		img := gocv.IMRead("gopher.png", gocv.IMReadUnchanged)
		defer img.Close()

		// borrando imagem
		imgBlured := gocv.NewMat()
		defer imgBlured.Close()
		start := time.Now()
		gocv.MedianBlur(img, &imgBlured, 7)
		elapsed := time.Since(start)
		log.Printf("median time: %s", elapsed)
		fmt.Println("Image blured")

		// fazendo as bordas
		go makeEdges(imgBlured, edgesChan)

		// aplicando filtro bilateral
		go filter(imgBlured, filterChan)

		// cartunizando
		go makeToon(img, edgesChan, filterChan, toonChan)
		window1 := gocv.NewWindow("original image")
		defer window1.Close()
		window2 := gocv.NewWindow("toonifyed image")
		defer window2.Close()

		window1.IMShow(img)
		window2.IMShow(<-toonChan)

		window1.WaitKey(0)

		doneChan <- "Done!"

	}()

	<-doneChan
	elapsed := time.Since(start)
	log.Printf("total time: %s", elapsed)
	close(doneChan)
	close(edgesChan)
	close(filterChan)
	close(toonChan)

}
