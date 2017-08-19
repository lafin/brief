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
	"github.com/steakknife/hamming"
	"github.com/tajtiattila/blur"
)

func getDescriptors(pixels []int, width int, keypoints []int, randomWindowOffsets []int, randomImageOffsets *map[int][]int, count int) []int {
	descriptors := make([]int, (len(keypoints)>>1)*(count>>5))
	offsets := getRandomOffsets(randomWindowOffsets, randomImageOffsets, width, count)
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
	confidence float32
}

func match(keypoints1, descriptors1, keypoints2, descriptors2 []int, count int) []Point {
	len1 := len(keypoints1) >> 1
	len2 := len(keypoints2) >> 1
	matches := make([]Point, len1)

	for i := 0; i < len1; i++ {
		min := math.MaxInt32
		minj := 0
		for j := 0; j < len2; j++ {
			dist := 0
			n := count >> 5
			for k := 0; k < n; k++ {
				dist += hamming.CountBitsInt(descriptors1[i*n+k] ^ descriptors2[j*n+k])
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
			confidence: 1.0 - float32(min)/float32(count),
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

func getRandomOffsets(randomWindowOffsets []int, randomImageOffsets *map[int][]int, width int, count int) []int {
	if _, ok := (*randomImageOffsets)[width]; !ok {
		imageOffsets := make([]int, 2*count)
		imagePosition := 0
		for i := 0; i < count; i++ {
			imageOffsets[imagePosition] = randomWindowOffsets[4*i]*width + randomWindowOffsets[4*i+1]
			imagePosition++
			imageOffsets[imagePosition] = randomWindowOffsets[4*i+2]*width + randomWindowOffsets[4*i+3]
			imagePosition++
		}
		(*randomImageOffsets)[width] = imageOffsets
	}

	return (*randomImageOffsets)[width]
}

func uniformRandom(a, b int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(b-a) + a
}

func toGray(path string) (*image.Gray, int, int) {
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

	src = blur.Gaussian(src, 2, blur.ReuseSrc)

	// Create a new grayscale image
	bounds := src.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	gray := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			oldColor := src.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor)
			gray.Set(x, y, grayColor)
		}
	}

	return gray, width, height
}

func grayImageToPixList(gray *image.Gray, width, height int) []int {
	pixList := make([]int, width*height)
	for index := 0; index < width*height; index++ {
		pixList[index] = int(gray.Pix[index])
	}

	return pixList
}

func main() {
	count := 256
	randomWindowOffsets := make([]int, 4*count)
	for i := 0; i < 4*count; i++ {
		randomWindowOffsets[i] = uniformRandom(-15, 16)
	}
	randomImageOffsets := make(map[int][]int)

	gray1, width1, height1 := toGray("image_1.png")
	gray2, width2, height2 := toGray("image_2.png")
	pixList1 := grayImageToPixList(gray1, width1, height1)
	pixList2 := grayImageToPixList(gray2, width2, height2)

	pixList1, pixList2 = getImages()
	width1, height1 = getSizes()
	width2, height2 = getSizes()

	corners1 := fast.FindCorners(pixList1, width1, height1, 40)
	descriptors1 := getDescriptors(pixList1, width1, corners1, randomWindowOffsets, &randomImageOffsets, count)
	corners2 := fast.FindCorners(pixList2, width2, height2, 40)
	descriptors2 := getDescriptors(pixList2, width2, corners2, randomWindowOffsets, &randomImageOffsets, count)

	matches := reciprocalMatch(corners1, descriptors1, corners2, descriptors2, count)
	for _, match := range matches {
		fmt.Println(match.index1, match.index2, match.keypoint1, match.keypoint2, match.confidence)
	}

	fmt.Println("done")
}
