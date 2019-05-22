package main

import (
  "gocv.io/x/gocv"
  "fmt"
)

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
  imgKmeans = imgKmeans.Reshape(-1, 3)
  //imageToonify =
  fmt.Println(imgKmeans)

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
