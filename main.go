package main

import (
	"fmt"
	"Toonify/kmeans"
	"gocv.io/x/gocv"
)

func imgRework(labels []int, centroids []kmeans.Observation, rows int, cols int)(gocv.Mat, error){
  mat := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV64F)
  for i:=0; i<rows; i++ {
    for j:=0; j<cols; j++{
        value := labels[i*cols + j]
        ch := centroids[value]
        mat.SetDoubleAt3(i, j, 0, ch[0])
        mat.SetDoubleAt3(i, j, 1, ch[1])
        mat.SetDoubleAt3(i, j, 2, ch[2])
    }
  }
  return mat, nil
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
	teste, _ := imgKmeans.DataPtrFloat64()
	//imgKmeans = imgKmeans.Reshape(3, img.Cols()*img.Rows())

	mat, _ := reshape(teste, img.Cols()*img.Rows(), 3)
	fmt.Println(mat)

	labels, centroids, _ := kmeans.Kmeans(mat, 3, kmeans.EuclideanDistance, 10)
  imgQuantized, _ := imgRework(labels, centroids, img.Rows(), img.Cols())

  img2 := gocv.NewMat()
  imgQuantized.ConvertTo(&img2, gocv.MatTypeCV32S)

  fmt.Println(img2)
	fmt.Println(len(labels))
  fmt.Println(centroids)
	fmt.Println(imgKmeans.Cols())
	fmt.Println(imgKmeans.Rows())
	fmt.Println(imgKmeans.Channels())
  fmt.Println(imgQuantized.Size())

	window := gocv.NewWindow("original gopher")
	window2 := gocv.NewWindow("gopher blured")
	window3 := gocv.NewWindow("gopher edges")
	window4 := gocv.NewWindow("gopher filtered")
	//window5 := gocv.NewWindow("gopher reshape")
	window.IMShow(img)
	window2.IMShow(imgBlured)
	window3.IMShow(imgEdges)
	window4.IMShow(imgFiltered)
	//window5.IMShow(imgKmeans)
	window.WaitKey(0)
}
