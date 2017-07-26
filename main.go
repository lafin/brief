package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func getDescriptors(pixels []int, width int, keypoints []int, count int) []int {
	descriptors := make([]int, (len(keypoints)>>1)*(count>>5))
	offsets := getRandomOffsets(width, count)
	descriptorWord := 0
	position := 0

	for i := 0; i < len(keypoints); i += 2 {
		var w = width*keypoints[i+1] + keypoints[i]

		var offsetsPosition = 0
		for j := 0; j < count; j++ {
			offsetsPosition += 2
			if pixels[offsets[offsetsPosition-1]+w] < pixels[offsets[offsetsPosition]+w] {
				descriptorWord |= 1 << (uint(j) & 31)
			}

			if ((j + 1) & 31) == 0 {
				position++
				descriptors[position] = descriptorWord
				descriptorWord = 0
			}
		}
	}

	return descriptors
}

// Point - struct of point
type Point struct {
	index1     int
	index2     int
	keypoint1  [2]int
	keypoint2  [2]int
	confidence int
}

func match(keypoints1, descriptors1, keypoints2, descriptors2 []int, count int) []Point {
	len1 := len(keypoints1) >> 1
	len2 := len(keypoints2) >> 1
	matches := make([]Point, len1)

	for i := 0; i < len1; i++ {
		min := math.MaxInt8
		minj := 0
		for j := 0; j < len2; j++ {
			dist := 0
			n := count >> 5
			for k := 0; k < n; k++ {
				dist += hammingWeight(descriptors1[i*n+k] ^ descriptors2[j*n+k])
			}
			if dist < min {
				min = dist
				minj = j
			}
		}
		matches[i] = Point{
			index1:     i,
			index2:     minj,
			keypoint1:  [2]int{keypoints1[2*i], keypoints1[2*i+1]},
			keypoint2:  [2]int{keypoints2[2*minj], keypoints2[2*minj+1]},
			confidence: 1 - min/count,
		}
	}

	return matches
}

func reciprocalMatch(keypoints1, descriptors1, keypoints2, descriptors2 []int, count int) []Point {
	var matches []Point
	if len(keypoints1) == 0 || len(keypoints2) == 0 {
		return matches
	}

	var matches1 = match(keypoints1, descriptors1, keypoints2, descriptors2, count)
	var matches2 = match(keypoints2, descriptors2, keypoints1, descriptors1, count)
	for i := 0; i < len(matches1); i++ {
		if matches2[matches1[i].index2].index2 == i {
			matches = append(matches, matches1[i])
		}
	}
	return matches
}

func getRandomOffsets(width int, count int) []int {
	var randomWindowOffsets []int
	if randomWindowOffsets == nil {
		var windowPosition = 0
		windowOffsets := make([]int, 4*count)
		for i := 0; i < count; i++ {
			windowPosition++
			windowOffsets[windowPosition] = uniformRandom(-15, 16)
			windowPosition++
			windowOffsets[windowPosition] = uniformRandom(-15, 16)
			windowPosition++
			windowOffsets[windowPosition] = uniformRandom(-15, 16)
			windowPosition++
			windowOffsets[windowPosition] = uniformRandom(-15, 16)
		}
		randomWindowOffsets = windowOffsets
	}

	randomImageOffsets := make(map[int][]int)
	if _, ok := randomImageOffsets[width]; !ok {
		var imagePosition = 0
		imageOffsets := make([]int, 2*count)
		for j := 0; j < count; j++ {
			imagePosition++
			imageOffsets[imagePosition] = randomWindowOffsets[4*j]*width + randomWindowOffsets[4*j+1]
			imagePosition++
			imageOffsets[imagePosition] = randomWindowOffsets[4*j+2]*width + randomWindowOffsets[4*j+3]
		}
		randomImageOffsets[width] = imageOffsets

	}

	return randomImageOffsets[width]
}

func hammingWeight(i int) int {
	i = i - ((i >> 1) & 0x55555555)
	i = (i & 0x33333333) + ((i >> 2) & 0x33333333)

	return ((i + (i>>4)&0xF0F0F0F) * 0x1010101) >> 24
}

func uniformRandom(a, b int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(b-a) + a
}

func round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

func main() {
	fmt.Println("done")
}
