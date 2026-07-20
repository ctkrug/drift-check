package ecosystem

import "strings"

// reconcile compares every pair of pins for the same ecosystem and reports
// whether they disagree. Versions are compared as dotted prefixes: "1.24"
// matches "1.24.3". When drift is found, detail names every source and its
// version so a multi-source mismatch is fully visible.
func reconcile(pins []Pin) (drift bool, detail string) {
	if len(pins) < 2 {
		return false, ""
	}
	for i := 0; i < len(pins)-1 && !drift; i++ {
		for j := i + 1; j < len(pins); j++ {
			if !versionsAgree(pins[i].Version, pins[j].Version) {
				drift = true
				break
			}
		}
	}
	if !drift {
		return false, ""
	}

	parts := make([]string, len(pins))
	for i, pin := range pins {
		parts[i] = pin.Source + " says " + pin.Version
	}
	return true, strings.Join(parts, ", ")
}

// versionsAgree reports whether two version strings are compatible,
// comparing only as many dotted components as the shorter one has.
func versionsAgree(a, b string) bool {
	partsA, partsB := strings.Split(a, "."), strings.Split(b, ".")
	length := len(partsA)
	if len(partsB) < length {
		length = len(partsB)
	}
	for i := 0; i < length; i++ {
		if partsA[i] != partsB[i] {
			return false
		}
	}
	return true
}
