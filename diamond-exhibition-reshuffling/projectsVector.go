package main

const numProjects = 21

// projectsVector is a vector of integers of length numProjects that supports standard vector operations.
type projectsVector [numProjects]int

// smul computes the scalar product of two vectors.
func (a projectsVector) smul(b projectsVector) int {
	var res int
	for i := range a {
		res += a[i] * b[i]
	}
	return res
}

// add computes the sum of two vectors.
func (a projectsVector) add(b projectsVector) projectsVector {
	var res projectsVector
	for i := range a {
		res[i] = a[i] + b[i]
	}
	return res
}

// mask returns a vector with 1s in the positions where the original vector is non-zero
func (a projectsVector) mask() projectsVector {
	var res projectsVector
	for i, x := range a {
		if x != 0 {
			res[i] = 1
		}
	}
	return res
}

// copy returns a copy of the vector.
func (a projectsVector) copy() projectsVector {
	var c projectsVector
	copy(c[:], a[:])
	return c
}
