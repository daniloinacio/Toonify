package main

import (
	"Toonify/kmeans"
	"fmt"
	"gocv.io/x/gocv"
	"log"
	"os"
	"time"
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

func main() {
	start := time.Now()

	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\t./toonify [image file]")
		return
	}

	img := gocv.IMRead(os.Args[1], gocv.IMReadUnchanged)
	if img.Empty() {
		fmt.Println("Couldn't open image")
		return
	}

	fmt.Println("Toonifying...")
	gocv.IMWrite("result/1original.jpg", img)
	defer img.Close()

	// borrando imagem
	imgBlured := gocv.NewMat()
	defer imgBlured.Close()
	gocv.MedianBlur(img, &imgBlured, 7)
	gocv.IMWrite("result/2blured.jpg", imgBlured)
	fmt.Println("Blured")

	// fazendo as bordas
	imgEdges := gocv.NewMat()
	defer imgEdges.Close()
	gocv.Canny(imgBlured, &imgEdges, 62.5, 125)
	gocv.BitwiseNot(imgEdges, &imgEdges)
	gocv.CvtColor(imgEdges, &imgEdges, gocv.ColorGrayToBGR)
	gocv.IMWrite("result/5edges.jpg", imgEdges)
	fmt.Println("Edges taken")

	// aplicando filtro bilateral
	imgFiltered := gocv.NewMat()
	defer imgFiltered.Close()
	gocv.BilateralFilter(imgBlured, &imgFiltered, 7, 35, 35)
	gocv.IMWrite("result/3bfilter.jpg", imgFiltered)
	fmt.Println("Filtered")

	// quantização da imagem
	data := formatData(imgFiltered)
	clusteredData, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
	imgQuantized, _ := imgRework(clusteredData, centroids, img.Rows(), img.Cols())
	defer imgQuantized.Close()
	gocv.IMWrite("result/4quantized.jpg", imgQuantized)
	fmt.Println("Quantized")

	// juntando as bordas com a quantizada
	imgToonify := gocv.NewMat()
	defer imgToonify.Close()
	gocv.BitwiseAnd(imgEdges, imgQuantized, &imgToonify)
	gocv.IMWrite("result/6final.jpg", imgToonify)
	fmt.Println("Toonified.")

	// mostrando o tempo total gasto
	elapsed := time.Since(start)
	log.Printf("total time: %s", elapsed)

	window1 := gocv.NewWindow("original image")
	defer window1.Close()
	window2 := gocv.NewWindow("toonifyed image")
	defer window2.Close()

	window1.IMShow(img)
	window2.IMShow(imgToonify)
	window1.WaitKey(0)
}
