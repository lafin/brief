/**
 * Brief intends for "Binary Robust Independent Elementary Features".This
 * method generates a binary string for each keypoint found by an extractor
 * method.
 */

package brief

import (
	"math"
	"math/rand"
	"time"

	"github.com/steakknife/hamming"
)

/**
 * Caches coordinates values of (x,y)-location pairs uniquely chosen during
 * the initialization.
 */
var randomImageOffsets = make(map[int][]int)

/**
 * The set of binary tests is defined by the nd (x,y)-location pairs
 * uniquely chosen during the initialization. Values could vary between N =
 * 128,256,512. N=128 yield good compromises between speed, storage
 * efficiency, and recognition rate.
 */
var count = 256

// Point - struct of point
type Point struct {
	index1     int
	index2     int
	keypoint1  [2]int
	keypoint2  [2]int
	confidence float32
}

func uniformRandom(a, b int) int {
	rand.Seed(time.Now().UTC().UnixNano())

	return rand.Intn(b-a) + a
}

/**
 * Gets the coordinates values of (x,y)-location pairs uniquely chosen
 * during the initialization.
 */
func getRandomOffsets(randomWindowOffsets []int, width int) []int {
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

/**
 * Matches sets of features {mi} and {m′j} extracted from two images taken
 * from similar, and often successive, viewpoints. A classical procedure
 * runs as follows. For each point {mi} in the first image, search in a
 * region of the second image around location {mi} for point {m′j}. The
 * search is based on the similarity of the local image windows, also known
 * as kernel windows, centered on the points, which strongly characterizes
 * the points when the images are sufficiently close. Once each keypoint is
 * described with its binary string, they need to be compared with the
 * closest matching point. Distance metric is critical to the performance of
 * in- trusion detection systems. Thus using binary strings reduces the size
 * of the descriptor and provides an interesting data structure that is fast
 * to operate whose similarity can be measured by the Hamming distance.
 */
func match(keypoints1, descriptors1, keypoints2, descriptors2 []int) []Point {
	len1 := len(keypoints1) >> 1
	len2 := len(keypoints2) >> 1
	matches := make([]Point, len1)

	for i := 0; i < len1; i++ {
		min := math.MaxInt32
		minj := 0
		for j := 0; j < len2; j++ {
			dist := 0
			// Optimizing divide by 32 operation using binary shift
			// (count >> 5) === count/32.
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

// InitOffsets - delta values of (x,y)-location pairs uniquely chosen during the initialization.
func InitOffsets() []int {
	randomWindowOffsets := make([]int, 4*count)
	for i := 0; i < 4*count; i++ {
		randomWindowOffsets[i] = uniformRandom(-15, 16)
	}

	return randomWindowOffsets
}

// GetDescriptors - Generates a binary string for each found keypoints extracted using an extractor method.
func GetDescriptors(pixels []int, width int, keypoints []int, randomWindowOffsets []int) []int {
	// Optimizing divide by 32 operation using binary shift
	// (count >> 5) === count/32.
	descriptors := make([]int, (len(keypoints)>>1)*(count>>5))
	offsets := getRandomOffsets(randomWindowOffsets, width)
	descriptorWord := 0
	position := 0

	for i := 0; i < len(keypoints); i += 2 {
		w := width*keypoints[i+1] + keypoints[i]

		offsetsPosition := 0
		for j := 0; j < count; j++ {
			offsetsPosition += 2
			if pixels[offsets[offsetsPosition-2]+w] < pixels[offsets[offsetsPosition-1]+w] {
				// The bit in the position `j % 32` of descriptorWord should be set to 1. We do
				// this by making an OR operation with a binary number that only has the bit
				// in that position set to 1. That binary number is obtained by shifting 1 left by
				// `j % 32` (which is the same as `j & 31` left) positions.
				descriptorWord |= 1 << (uint(j) & 31)
			}

			// If the next j is a multiple of 32, we will need to use a new descriptor word to hold
			// the next results.
			if ((j + 1) & 31) == 0 {
				descriptors[position] = descriptorWord
				position++
				descriptorWord = 0
			}
		}
	}

	return descriptors
}

// ReciprocalMatch - Removes matches outliers by testing matches on both directions.
func ReciprocalMatch(keypoints1, descriptors1, keypoints2, descriptors2 []int) []Point {
	var matches []Point
	if len(keypoints1) == 0 || len(keypoints2) == 0 {
		return matches
	}

	var matches1 = match(keypoints1, descriptors1, keypoints2, descriptors2)
	var matches2 = match(keypoints2, descriptors2, keypoints1, descriptors1)

	for i := 0; i < len(matches1); i++ {
		if matches2[matches1[i].index2].index2 == i {
			matches = append(matches, matches1[i])
		}
	}

	return matches
}
