package main

import (
	"Toonify/kmeans"
	"fmt"
	"log"
	"os"
	"time"

	"gocv.io/x/gocv"
)

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
	data := kmeans.FormatData(imgFiltered)
	clusteredData, centroids, _ := kmeans.Kmeans(data, 24, kmeans.EuclideanDistance, 10)
	imgQuantized, _ := kmeans.ImgRework(clusteredData, centroids, img.Rows(), img.Cols())
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
