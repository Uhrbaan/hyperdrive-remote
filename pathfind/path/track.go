package path

import "strings"

// implementing necessary functions for sorting
type ByDetectedProbability []string

func (p ByDetectedProbability) Len() int      { return len(p) }
func (p ByDetectedProbability) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ByDetectedProbability) Less(i, j int) bool {
	// Estimated probabilitiy that the section does *not* get detected
	probability := func(s string) float32 {
		switch true {
		case strings.Contains(s, "curve.inner"):
			return 0.3
		case strings.Contains(s, "intersection.left"):
			return 0.2
		case strings.Contains(s, "intersection.right"):
			return 0.2
		case strings.Contains(s, "intersection.bottom"):
			return 0.5
		case strings.Contains(s, "curve.outer"):
			return 0.6
		case strings.Contains(s, "straight"):
			return 0.8
		default:
			return 0
		}
	}

	pI := probability(p[i])
	pJ := probability(p[j])

	return pI < pJ
}
