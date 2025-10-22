package generic

import "reflect"

// ReturnZero generate the zero value of a generic type
func ReturnZero[T any](T) T {
	var zero T
	return zero
}

// Contains check if an array contains a given value
func Contains[T comparable](elems []T, v T) bool {
	// T type parameter with the comparable constraint.
	// It is a built-in constraint that describes any type whose values can be compared,
	// i.e., we can use == and != operators on them.
	for _, s := range elems {
		if v == s {
			return true
		}
	}

	return false
}

// Remove removes all occurrence of the given element
func Remove[T comparable](s []T, val T) []T {
	for i, v := range s {
		if v == val {
			s = append(s[:i], s[i+1:]...)
		}
	}

	return s
}

// RemoveDuplicates removes the occurrence of duplicate elements in an array.
func RemoveDuplicates[T comparable](array []T) []T {
	exist := make(map[T]struct{})
	var result []T

	for _, val := range array {
		if _, ok := exist[val]; !ok {
			result = append(result, val)
			exist[val] = struct{}{}
		}
	}

	return result
}

// IsZero checks if the given value v is the zero value for its type
func IsZero[T any](v T) bool {
	// check if the type of v is nil
	if reflect.TypeOf(v) == nil {
		return true
	}

	// compare v to its zero value
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
