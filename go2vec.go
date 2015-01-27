package go2vec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

type WordDistance struct {
	Word     string
	Distance float32
}

type Vector []float32

func ReadVectors(r *bufio.Reader) (map[string]Vector, error) {
	var nWords uint64
	_, err := fmt.Fscanf(r, "%d", &nWords)
	if err != nil {
		return nil, err
	}

	var vSize uint64
	_, err = fmt.Fscanf(r, "%d", &vSize)
	if err != nil {
		return nil, err
	}

	vecs := make(map[string]Vector)

	for w := uint64(0); w < nWords; w++ {
		word, err := r.ReadString(' ')
		word = strings.TrimSpace(word)
		vec := make([]float32, vSize)

		err = binary.Read(r, binary.LittleEndian, vec)
		if err != nil {
			return nil, err
		}

		normalize(vec)

		vecs[word] = vec
	}

	return vecs, nil
}

func main() {
	f, err := os.Open("vectors.bin")
	if err != nil {
		log.Fatal(err)
	}

	vecs, err := ReadVectors(bufio.NewReader(io.Reader(f)))
	if err != nil {
		log.Fatal(err)
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)

		results, err := Distance(vecs, line, 10)

		for _, wordDistance := range results {
			fmt.Println(wordDistance.Word, wordDistance.Distance)
		}
	}
}

func dotProduct(v, w []float32) float32 {
	sum := float32(0)

	for idx, val := range v {
		sum += val * w[idx]
	}

	return sum
}

func Analogy(vecs map[string]Vector, word1, word2, word3 string, limit int) ([]WordDistance, error) {
	v1, ok := vecs[word1]
	if !ok {
		return nil, fmt.Errorf("Unknown word: %s", word1)
	}

	v2, ok := vecs[word2]
	if !ok {
		return nil, fmt.Errorf("Unknown word: %s", word2)
	}

	v3, ok := vecs[word3]
	if !ok {
		return nil, fmt.Errorf("Unknown word: %s", word3)
	}

	v4 := plus(minus(v2, v1), v3)

	skips := map[string]interface{}{
		word1: nil,
		word2: nil,
		word3: nil,
	}

	return distance(vecs, v4, skips, limit)
}

func Distance(vecs map[string]Vector, word string, limit int) ([]WordDistance, error) {
	v, ok := vecs[word]
	if !ok {
		return nil, fmt.Errorf("Unknown word: %s", word)
	}

	skips := map[string]interface{}{
		word: nil,
	}

	return distance(vecs, v, skips, limit)
}

func distance(vecs map[string]Vector, vec Vector, skips map[string]interface{}, limit int) ([]WordDistance, error) {
	results := make([]WordDistance, 0)

	for vecWord, w := range vecs {
		// Skip words in the skip set.
		_, ok := skips[vecWord]
		if ok {
			continue
		}

		dist := dotProduct(vec, w)

		ip := sort.Search(len(results), func(i int) bool {
			return results[i].Distance <= dist
		})
		if ip < limit {
			results = insertWithLimit(results, limit, ip, WordDistance{vecWord, dist})
		}
	}

	return results, nil
}

func insertWithLimit(slice []WordDistance, limit, index int, value WordDistance) []WordDistance {
	if len(slice) < limit {
		slice = append(slice, WordDistance{})
	}

	copy(slice[index+1:], slice[index:len(slice)-1])
	slice[index] = value
	return slice
}

func minus(v, w []float32) []float32 {
	result := make([]float32, len(v))

	for idx, val := range v {
		result[idx] = val - w[idx]
	}

	return result
}

func normalize(vec []float32) {
	// Normalize
	vecLen := float32(0)
	for _, val := range vec {
		vecLen += val * val
	}

	vecLen = float32(math.Sqrt(float64(vecLen)))

	for idx, val := range vec {
		vec[idx] = val / vecLen
	}
}

func plus(v, w []float32) []float32 {
	result := make([]float32, len(v))

	for idx, val := range v {
		result[idx] = val + w[idx]
	}

	return result
}
