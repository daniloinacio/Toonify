package main

import (
	"fmt"
	"Toonify/kmeans"
	"gocv.io/x/gocv"
)

//Reconstroi a imagem a partir da clusterização do kmeans
func imgRework(mat *gocv.Mat, labels []int, centroids []kmeans.Observation, rows int, cols int)( error){
  //mat := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
  for i:=0; i<rows; i++ {
    for j:=0; j<cols; j++{
        value := labels[i*cols + j]
				//valores do pixel na posição i, j nos 3 canais
        ch := centroids[value]
        mat.SetUCharAt3(i, j, 0, uint8(ch[0]))
				//os valores 37 e 74 são o offset da distorção da função Set do gocv
        mat.SetUCharAt3(i, j-37, 1, uint8(ch[1]))
				mat.SetUCharAt3(i, j-74, 2, uint8(ch[2]))

		}
  }
  return nil
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
  //fmt.Println(img.GetUCharAt3(0, 0, 1))
	/*for i:=0; i<50; i++{
			for j:=0; j<37; j++{
				img.SetUCharAt3(100+i, j, 0, uint8(0))
				img.SetUCharAt3(100+i, j, 1, uint8(0))
				img.SetUCharAt3(100+i, j, 2, uint8(0))
			}
	}
	rgb := gocv.Split(img)
	*/

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
	teste, _ := imgKmeans.DataPtrFloat64()
	//imgKmeans = imgKmeans.Reshape(3, img.Cols()*img.Rows())

	mat, _ := reshape(teste, img.Cols()*img.Rows(), 3)
	fmt.Println(img.Rows(), img.Cols())

	labels, centroids, _ := kmeans.Kmeans(mat, 24, kmeans.EuclideanDistance, 10)
  imgRework(&img, labels, centroids, img.Rows(), img.Cols())
	imgToonify := gocv.NewMat()
	gocv.BitwiseAnd(imgEdges, img, &imgToonify)

  //fmt.Println(centroids)
	//window := gocv.NewWindow("original gopher")
	window2 := gocv.NewWindow("gopher blured")
	window3 := gocv.NewWindow("gopher edges")
	window4 := gocv.NewWindow("gopher filtered")
  window5 := gocv.NewWindow("gopher quantized")
	window6 := gocv.NewWindow("gopher toonifyed")
	//window6 := gocv.NewWindow("B")
	//window7 := gocv.NewWindow("G")
	//window8 := gocv.NewWindow("R")
	//window.IMShow(img)
	window2.IMShow(imgBlured)
	window3.IMShow(imgEdges)
	window4.IMShow(imgFiltered)
  window5.IMShow(img)
	window6.IMShow(imgToonify)
	//window6.IMShow(rgb[0])
	//window7.IMShow(rgb[1])
	//window8.IMShow(rgb[2])

	window2.WaitKey(0)
}
