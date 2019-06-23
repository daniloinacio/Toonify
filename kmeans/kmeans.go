package kmeans

import (
	"math"
	"math/rand"

	"gocv.io/x/gocv"
)

// Pixel: Abstração de uma slice N-dimensional de float64
type Pixel []float64

// ClusteredPixel: Abstração de pixel com número de cluster
type ClusteredPixel struct {
	ClusterNumber int
	Pixel
}

// Soma dois pixeis
func (Pixel Pixel) Add(otherPixel Pixel) {
	for i, j := range otherPixel {
		Pixel[i] += j
	}
}

// Multiplica um pixel por um escalar
func (Pixel Pixel) Mul(scalar float64) {
	for i := range Pixel {
		Pixel[i] *= scalar
	}
}

// Distance Function: Calcula a distância entre pixeis
type DistanceFunction func(first, second []float64) (float64, error)

func ManhattanDistance(firstVector, secondVector []float64) (float64, error) {
	distance := 0.
	for ii := range firstVector {
		distance += math.Abs(firstVector[ii] - secondVector[ii])
	}
	return distance, nil
}

func EuclideanDistance(firstVector, secondVector []float64) (float64, error) {
	distance := 0.
	for ii := range firstVector {
		distance += (firstVector[ii] - secondVector[ii]) * (firstVector[ii] - secondVector[ii])
	}
	return math.Sqrt(distance), nil
}

//Reconstroi a imagem a partir da clusterização do kmeans
func ImgRework(clusteredData []ClusteredPixel, centroids []Pixel, rows int, cols int) (gocv.Mat, error) {
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
func FormatData(img gocv.Mat) []ClusteredPixel {
	imgFloat64 := gocv.NewMat()
	defer imgFloat64.Close()
	img.ConvertTo(&imgFloat64, gocv.MatTypeCV64F)
	slice, _ := imgFloat64.DataPtrFloat64()

	rows := img.Rows() * img.Cols()
	cols := img.Channels()

	data := make([]ClusteredPixel, rows)
	for i := 0; i < rows; i++ {
		aux := make([]float64, cols)
		for j := 0; j < cols; j++ {
			aux[j] = slice[i*cols+j]
		}
		data[i].Pixel = aux
	}
	return data
}

// Encontra o indice do centroid/cluster mais proximo do pixel
func near(p ClusteredPixel, centroid []Pixel, distanceFunction DistanceFunction) (int, float64) {
	indexOfCluster := 0
	minSquaredDistance, _ := distanceFunction(p.Pixel, centroid[0])
	for i := 1; i < len(centroid); i++ {
		squaredDistance, _ := distanceFunction(p.Pixel, centroid[i])
		if squaredDistance < minSquaredDistance {
			minSquaredDistance = squaredDistance
			indexOfCluster = i
		}
	}
	return indexOfCluster, math.Sqrt(minSquaredDistance)
}

// Retorna uma slice de k centroids inicializados com
// valores de pixeis escolhidos aleatoriamente do data
func Seed(data []ClusteredPixel, k int) []Pixel {
	s := make([]Pixel, k)
	for i := 0; i < k; i++ {
		s[i] = data[rand.Intn(len(data))].Pixel
	}
	return s
}

// Implementa o algoritimo de clusterização K-means
func Kmeans(data []ClusteredPixel, centroid []Pixel, distanceFunction DistanceFunction, threshold int) ([]ClusteredPixel, []Pixel, error) {
	//centroid := seed(data, k)
	counter := 0

	mLen := make([]int, len(centroid))
	n := len(data[0].Pixel)
	newcentroid := make([]Pixel, len(centroid))
	for i := range centroid {
		newcentroid[i] = make(Pixel, n)
	}
	for {
		var changes int
		// Atualiza a clusterização dos pixeis
		for i, p := range data {

			closestCluster, _ := near(p, centroid, distanceFunction)
			if closestCluster != p.ClusterNumber {
				changes++
				data[i].ClusterNumber = closestCluster
			}
			newcentroid[closestCluster].Add(p.Pixel)
			mLen[closestCluster]++
		}
		// Calcula os novos centroids
		for i := range centroid {
			newcentroid[i].Mul(1 / float64(mLen[i]))
			centroid[i] = newcentroid[i]
			newcentroid[i] = make(Pixel, n)
			mLen[i] = 0
		}

		counter++
		if changes == 0 || counter > threshold {
			break
		}

	}
	return data, centroid, nil
}
