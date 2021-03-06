package som

import (
	"container/heap"
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// Distance calculates metric distance between vectors a and b.
// If unsupported metric is requested Distance returns euclidean distance.
// It returns error if the supplied vectors are either nil or have different dimensions
func Distance(metric string, a, b []float64) (float64, error) {
	if a == nil || b == nil {
		return 0.0, fmt.Errorf("invalid vectors supplied. a: %v, b: %v", a, b)
	}
	if len(a) != len(b) {
		return 0.0, fmt.Errorf("Incorrect vector dims. a: %d, b: %d", len(a), len(b))
	}

	switch metric {
	case "euclidean":
		return euclideanVec(a, b), nil
	default:
		return euclideanVec(a, b), nil
	}
}

// DistanceMx calculates metric distance matrix for the supplied matrix.
// Distance matrix is also known in literature as dissimilarity matrix.
// DistanceMx returns a hollow symmetric matrix where an item x_ij contains the distance between
// vectors stored in rows i and j.  If an unknown metric is supplied Euclidean distance is computed.
// It returns error if the supplied matrix is nil.
func DistanceMx(metric string, m *mat.Dense) (*mat.Dense, error) {
	if m == nil {
		return nil, fmt.Errorf("invalid matrix supplied: %v", m)
	}

	switch metric {
	case "euclidean":
		return euclideanMx(m), nil
	default:
		return euclideanMx(m), nil
	}
}

// ClosestVec finds the closest vector to v in the list of vectors stored in m rows
// using the supplied distance metric. It returns an index to matrix m rows.
// If unsupported metric is requested, ClosestVec falls over to euclidean metric.
// If several vectors of the same distance are found, it returns the index of the first one found.
// ClosestVec returns error if either v or m are nil or if the v dimension is different from
// the number of m columns. When the ClosestVec fails with error returned index is set to -1.
func ClosestVec(metric string, v []float64, m *mat.Dense) (int, error) {
	// vector can't be nil
	if v == nil || len(v) == 0 {
		return -1, fmt.Errorf("invalid vector: %v", v)
	}
	// matrix cant be nil
	if m == nil {
		return -1, fmt.Errorf("invalid matrix: %v", m)
	}
	// check if the dimensions are ok
	rows, _ := m.Dims()
	closest := 0
	dist := math.MaxFloat64
	for i := 0; i < rows; i++ {
		d, err := Distance(metric, v, m.RawRowView(i))
		if err != nil {
			return -1, err
		}
		if d < dist {
			dist = d
			closest = i
		}
	}

	return closest, nil
}

// ClosestNVec finds the N closest vectors to v in the list of vectors stored in m rows
// using the supplied distance metric. It returns a slice which contains indices to the m
// rows. The length of the slice is the same as number of requested closest vectors - n.
// ClosestNVec fails in the same way as ClosestVec. If n is higher than the number of
// rows in m, or if it is not a positive integer, it fails with error too.
func ClosestNVec(metric string, n int, v []float64, m *mat.Dense) ([]int, error) {
	// vector can't be nil
	if v == nil || len(v) == 0 {
		return nil, fmt.Errorf("invalid vector: %v", v)
	}
	// matrix cant be nil
	if m == nil {
		return nil, fmt.Errorf("invalid matrix: %v", m)
	}
	rows, _ := m.Dims()
	// n must be positive integer and smaller than number of rows in m
	if n <= 0 || n > rows {
		return nil, fmt.Errorf("invalid number of closest vectors requested: %d", n)
	}
	// we return slice of indices
	closest := make([]int, n)

	switch {
	case n == 1:
		idx, err := ClosestVec(metric, v, m)
		if err != nil {
			return nil, err
		}
		closest[0] = idx
	default:
		// no need to check for error
		h, _ := newFloat64Heap(n)
		rows, _ := m.Dims()
		for i := 0; i < rows; i++ {
			d, err := Distance(metric, v, m.RawRowView(i))
			if err != nil {
				return nil, err
			}
			f := &float64Item{val: d, index: i}
			heap.Push(h, f)
		}
		// add max vals to closest slide
		for j := 0; j < n; j++ {
			closest[j] = (heap.Pop(h).(*float64Item)).index
		}
	}

	return closest, nil
}

// BMUs returns a slice which contains indices of Best Match Unit (BMU) codebook vectors for each
// vector stored in data rows. Each item in the returned slice correspnds to index of BMU for
// a particular data sample. If some data row has more than one BMU the index of the first one found is used.
// It returns error if either the data or codebook are nil or if their dimensions are mismatched.
func BMUs(data, codebook *mat.Dense) ([]int, error) {
	// data can't be nil
	if data == nil {
		return nil, fmt.Errorf("invalid data supplied: %v", data)
	}
	// codebook cant be nil
	if codebook == nil {
		return nil, fmt.Errorf("invalid codebook supplied: %v", codebook)
	}

	rows, _ := data.Dims()
	bmus := make([]int, rows)
	// loop through all data
	for i := 0; i < rows; i++ {
		idx, err := ClosestVec("euclidean", data.RawRowView(i), codebook)
		if err != nil {
			return nil, err
		}
		bmus[i] = idx
	}

	return bmus, nil
}

// euclideanVec computes euclidean distance between vectors a and b.
func euclideanVec(a, b []float64) float64 {
	d := 0.0
	for i := 0; i < len(a); i++ {
		d += (a[i] - b[i]) * (a[i] - b[i])
	}

	return math.Sqrt(d)
}

// euclideanMx computes a matrix of euclidean distances between each row in m
func euclideanMx(m *mat.Dense) *mat.Dense {
	rows, _ := m.Dims()
	out := mat.NewDense(rows, rows, nil)

	for row := 0; row < rows-1; row++ {
		dist := 0.0
		a := m.RawRowView(row)
		for i := row + 1; i < rows; i++ {
			if i != row {
				b := m.RawRowView(i)
				dist = euclideanVec(a, b)
				out.Set(row, i, dist)
				out.Set(i, row, dist)
			}
		}
	}

	return out
}
