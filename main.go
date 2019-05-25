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

func main() {
	img := gocv.IMRead("gopher.png", gocv.IMReadUnchanged)

	imgBlured := gocv.NewMat()
	gocv.MedianBlur(img, &imgBlured, 7)

	imgEdges := gocv.NewMat()
	gocv.Canny(imgBlured, &imgEdges, 62.5, 125)
	gocv.BitwiseNot(imgEdges, &imgEdges)
	gocv.CvtColor(imgEdges, &imgEdges, gocv.ColorGrayToBGR)

	imgFiltered := gocv.NewMat()
	gocv.BilateralFilter(imgBlured, &imgFiltered, 7, 35, 35)

	imgKmeans := gocv.NewMat()
	imgFiltered.ConvertTo(&imgKmeans, gocv.MatTypeCV64F)
	imgFloat64, _ := imgKmeans.DataPtrFloat64()

	data, _ := reshape(imgFloat64, img.Cols()*img.Rows(), 3)
	fmt.Println(img.Rows(), img.Cols())

	labels, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
	imgQuantized, _ := imgRework(labels, centroids, img.Rows(), img.Cols())
	imgToonify := gocv.NewMat()

	gocv.BitwiseAnd(imgEdges, imgQuantized, &imgToonify)

	window := gocv.NewWindow("original gopher")
	window2 := gocv.NewWindow("gopher blured")
	window3 := gocv.NewWindow("gopher edges")
	window4 := gocv.NewWindow("gopher filtered")
	window5 := gocv.NewWindow("gopher quantized")
	window6 := gocv.NewWindow("gopher toonifyed")

	window.IMShow(img)
	window2.IMShow(imgBlured)
	window3.IMShow(imgEdges)
	window4.IMShow(imgFiltered)
	window5.IMShow(imgQuantized)
	window6.IMShow(imgToonify)

	window.WaitKey(0)
}
