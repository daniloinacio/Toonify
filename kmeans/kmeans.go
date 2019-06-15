package kmeans

import (
	"math"
	"math/rand"
)

// Observation: Data Abstraction for an N-dimensional
// observation
type Observation []float64

// Abstracts the Observation with a cluster number
// Update and computeation becomes more efficient
type ClusteredObservation struct {
	ClusterNumber int
	Observation
}

// Distance Function: To compute the distanfe between observations
type DistanceFunction func(first, second []float64) (float64, error)

/*
func (observation Observation) Sqd(otherObservation Observation) (ssq float64) {
	for i, j := range observation {
		d := j - otherObservation[i]
		ssq += d * d
	}
	return ssq
}
*/

// Summation of two vectors
func (observation Observation) Add(otherObservation Observation) {
	for i, j := range otherObservation {
		observation[i] += j
	}
}

// Multiplication of a vector with a scalar
func (observation Observation) Mul(scalar float64) {
	for i := range observation {
		observation[i] *= scalar
	}
}

// Dot Product of Two vectors
func (observation Observation) InnerProduct(otherObservation Observation) {
	for i := range observation {
		observation[i] *= otherObservation[i]
	}
}

// Outer Product of two arrays
// TODO: Need to be tested
func (observation Observation) OuterProduct(otherObservation Observation) [][]float64 {
	result := make([][]float64, len(observation))
	for i := range result {
		result[i] = make([]float64, len(otherObservation))
	}
	for i := range result {
		for j := range result[i] {
			result[i][j] = observation[i] * otherObservation[j]
		}
	}
	return result
}

// Find the closest observation and return the distance
// Index of observation, distance
func near(p ClusteredObservation, mean []Observation, distanceFunction DistanceFunction) (int, float64) {
	indexOfCluster := 0
	minSquaredDistance, _ := distanceFunction(p.Observation, mean[0])
	for i := 1; i < len(mean); i++ {
		squaredDistance, _ := distanceFunction(p.Observation, mean[i])
		if squaredDistance < minSquaredDistance {
			minSquaredDistance = squaredDistance
			indexOfCluster = i
		}
	}
	return indexOfCluster, math.Sqrt(minSquaredDistance)
}

// Instead of initializing randomly the seeds, make a sound decision of initializing
func seed(data []ClusteredObservation, k int) []Observation {
	s := make([]Observation, k)
	for i := 0; i < k; i++ {
		s[i] = data[rand.Intn(len(data))].Observation
	}
	/*d2 := make([]float64, len(data))
	for i := 1; i < k; i++ {
		var sum float64
		for j, p := range data {
			_, dMin := near(p, s[:i], distanceFunction)
			d2[j] = dMin * dMin
			sum += d2[j]
		}
		target := rand.Float64() * sum
		j := 0
		for sum = d2[0]; sum < target; sum += d2[j] {
			j++
		}
		s[i] = data[j].Observation
	}*/
	return s
}

// K-Means Algorithm with smart seeds
// as known as K-Means ++
func Kmeans(data []ClusteredObservation, k int, distanceFunction DistanceFunction, threshold int) ([]ClusteredObservation, []Observation, error) {
	mean := seed(data, k)
	counter := 0

	mLen := make([]int, len(mean))
	n := len(data[0].Observation)
	newMean := make([]Observation, len(mean))
	for i := range mean {
		newMean[i] = make(Observation, n)
	}
	for {
		var changes int
		for i, p := range data {

			closestCluster, _ := near(p, mean, distanceFunction)

			if closestCluster != p.ClusterNumber {
				changes++
				data[i].ClusterNumber = closestCluster
			}
			newMean[closestCluster].Add(p.Observation)
			mLen[closestCluster]++
		}
		for i := range mean {
			newMean[i].Mul(1 / float64(mLen[i]))
			mean[i] = newMean[i]
			newMean[i] = make(Observation, n)
			mLen[i] = 0
		}

		counter++
		if changes == 0 || counter > threshold {
			break
		}

	}
	return data, mean, nil
}
