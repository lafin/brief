package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png" // register the PNG format with the image package
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/lafin/fast"
)

func getDescriptors(pixels []int, width int, keypoints []int, count int) []int {
	descriptors := make([]int, (len(keypoints)>>1)*(count>>5))
	offsets := getRandomOffsets(width, count)
	descriptorWord := 0
	position := 0

	for i := 0; i < len(keypoints); i += 2 {
		w := width*keypoints[i+1] + keypoints[i]

		offsetsPosition := 0
		for j := 0; j < count; j++ {
			offsetsPosition += 2
			if pixels[offsets[offsetsPosition-2]+w] < pixels[offsets[offsetsPosition-1]+w] {
				descriptorWord |= 1 << (uint(j) & 31)
			}

			if ((j + 1) & 31) == 0 {
				descriptors[position] = descriptorWord
				position++
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
	randomWindowOffsets := make([]int, 4*count)
	for i := 0; i < 4*count; i++ {
		randomWindowOffsets[i] = uniformRandom(-15, 16)
	}

	randomImageOffsets := make(map[int][]int)
	if _, ok := randomImageOffsets[width]; !ok {
		imageOffsets := make([]int, 2*count)
		imagePosition := 0
		for i := 0; i < count; i++ {
			imageOffsets[imagePosition] = randomWindowOffsets[4*i]*width + randomWindowOffsets[4*i+1]
			imagePosition++
			imageOffsets[imagePosition] = randomWindowOffsets[4*i+2]*width + randomWindowOffsets[4*i+3]
			imagePosition++
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

func toGray(path string) ([]int, int, int) {
	infile, err := os.Open(path)
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	defer infile.Close()

	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	if err != nil {
		// replace this with real error handling
		panic(err)
	}

	// Create a new grayscale image
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	gray := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := src.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor)
			gray.Set(x, y, grayColor)
		}
	}

	pixList := make([]int, w*h)
	for index := 0; index < w*h; index++ {
		pixList[index] = int(gray.Pix[index])
	}

	return pixList, w, h
}

func main() {
	pixList1, width1, height1 := toGray("image_1.png")
	corners1 := fast.FindCorners(pixList1, width1, height1, 20)
	descriptors1 := getDescriptors(pixList1, width1, corners1, 512)

	pixList2, width2, height2 := toGray("image_2.png")
	corners2 := fast.FindCorners(pixList2, width2, height2, 20)
	descriptors2 := getDescriptors(pixList2, width2, corners2, 512)

	matches := reciprocalMatch(corners1, descriptors1, corners2, descriptors2, 512)
	fmt.Println(matches)

	fmt.Println("done")
}
