package main

import (
  "gocv.io/x/gocv"
  "fmt"
  "github.com/bugra/kmeans"
)

func reshape(slice []float64, rows int, cols int) ([][]float64, error){
  mat := make([][]float64, rows)
      for i:=0; i<rows; i++ {
        aux := make([]float64, cols)
        for j:=0; j<cols; j++ {
          aux[j] = slice[i*cols + j]
        }
        mat[i] = aux
      }
  return mat, nil
}


func main(){
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

  mat, _:= reshape(teste, img.Cols()*img.Rows(), 3)
  fmt.Println(mat)

  labels, _ := kmeans.Kmeans(mat, 3, kmeans.EuclideanDistance, 10)
  fmt.Println(labels)

  fmt.Println(imgKmeans.Cols())
  fmt.Println(imgKmeans.Rows())
  fmt.Println(imgKmeans.Channels())

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
