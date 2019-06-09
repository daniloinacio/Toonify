package main

import (
	"Toonify/kmeans"
	"fmt"
	"gocv.io/x/gocv"
)

//Reconstroi a imagem a partir da clusterização do kmeans
func imgRework(labels []int, centroids []kmeans.Observation, rows int, cols int) (gocv.Mat, error) {
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
			value := labels[i*cols+j]
			//valores do pixel na posição i, j nos 3 canais
			ch := centroids[value]
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

func reshape(slice []float64, rows int, cols int) ([][]float64, error) {
	mat := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		aux := make([]float64, cols)
		for j := 0; j < cols; j++ {
			aux[j] = slice[i*cols+j]
		}
		mat[i] = aux
	}
	return mat, nil
}

func makeEdges(imgBlured gocv.Mat, edgesChan chan gocv.Mat) {
	imgEdges := gocv.NewMat()
	gocv.Canny(imgBlured, &imgEdges, 62.5, 125)
	gocv.BitwiseNot(imgEdges, &imgEdges)
	gocv.CvtColor(imgEdges, &imgEdges, gocv.ColorGrayToBGR)
	fmt.Println("Image edged")
	edgesChan <- imgEdges
}

func filter(imgBlured gocv.Mat, filterChan chan gocv.Mat) {
	imgFiltered := gocv.NewMat()
	gocv.BilateralFilter(imgBlured, &imgFiltered, 7, 35, 35)
	fmt.Println("Image filtered")
	filterChan <- imgFiltered
}

func makeToon(img gocv.Mat, edgesChan chan gocv.Mat, filterChan chan gocv.Mat, toonChan chan gocv.Mat) {
	imgKmeans := gocv.NewMat()
	defer imgKmeans.Close()

	filteredImg := <-filterChan
	defer filteredImg.Close()

	filteredImg.ConvertTo(&imgKmeans, gocv.MatTypeCV64F)
	imgFloat64, _ := imgKmeans.DataPtrFloat64()

	data, _ := reshape(imgFloat64, img.Cols()*img.Rows(), 3)

	labels, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
	imgQuantized, _ := imgRework(labels, centroids, img.Rows(), img.Cols())
	defer imgQuantized.Close()

	imgToonify := gocv.NewMat()
	gocv.BitwiseAnd(<-edgesChan, imgQuantized, &imgToonify)
	fmt.Println("Image Toonified")
	toonChan <- imgToonify
}

func main() {
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

		gocv.MedianBlur(img, &imgBlured, 7)
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
	close(doneChan)
	close(edgesChan)
	close(filterChan)
	close(toonChan)
}
