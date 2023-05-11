package main

const numProjects = 21

// projectsVector is a vector of integers of length numProjects that supports standard vector operations.
type projectsVector [numProjects]int

// smul computes the scalar product of two vectors.
func (v projectsVector) smul(w projectsVector) int {
	var res int
	for i := range v {
		res += v[i] * w[i]
	}
	return res
}

// add computes the sum of two vectors.
func (v projectsVector) add(w projectsVector) projectsVector {
	var res projectsVector
	for i := range v {
		res[i] = v[i] + w[i]
	}
	return res
}

// sum computes the sum of all elements in the vector.
func (v projectsVector) sum() int {
	var s int
	for _, x := range v {
		s += x
	}
	return s
}

// asMask returns a vector with 1s in the positions where the original vector is non-zero
func (v projectsVector) asMask() projectsVector {
	var res projectsVector
	for i, x := range v {
		if x != 0 {
			res[i] = 1
		}
	}
	return res
}

// copy returns a copy of the vector.
func (v projectsVector) copy() projectsVector {
	var cp projectsVector
	copy(cp[:], v[:])
	return cp
}

// normalised returns the vector normalised to sum to 1.
func (v projectsVector) normalised() [numProjects]float64 {
	sum := float64(v.sum())

	var res [numProjects]float64
	for i, x := range v {
		res[i] = float64(x) / sum
	}
	return res
}
