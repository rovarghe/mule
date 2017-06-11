package plugin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//Version has components Major, Minor and Patch which are numbers followed by an optional hyphen and label
//1.2.3-alpha, 1.2.3-pre-alpha, 0.1.1-build-13013-alpha, etc
type Version struct {
	Major int
	Minor int
	Patch int
	Label string
}

// Compare one version to another
func (v Version) Compare(other Version) int {
	switch {
	case v.Major < other.Major:
		return -1
	case v.Major > other.Major:
		return 1
	default:
		switch {
		case v.Minor < other.Minor:
			return -1
		case v.Minor > other.Minor:
			return 1
		default:
			switch {
			case v.Patch < other.Patch:
				return -1
			case v.Patch > other.Patch:
				return 1
			default:
				return strings.Compare(v.Label, other.Label)
			}
		}

	}
}

// String representation of Version
func (v Version) String() string {
	if v.Label == "" {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Label)

}

// Equals compares two versions
func (v Version) Equals(o Version) bool {
	return v.Compare(o) == 0
}

// ParseVersion takes  a string representation and converts it into a Version type
// Value of 'error' will be nil if parse is successful
func ParseVersion(s string) (*Version, error) {
	var v Version
	i := strings.Index(s, ".")
	if i < 0 {
		i = len(s)
	}

	j, e := strconv.Atoi(s[:i])
	if e != nil {
		return nil, e
	}
	v.Major = j
	if i < len(s) {
		s = s[i+1:]
		i := strings.Index(s, ".")
		if i < 0 {
			i = len(s)
		}
		j, e := strconv.Atoi(s[:i])
		if e != nil {
			return nil, e
		}
		v.Minor = j

		if i < len(s) {
			s = s[i+1:]
			i := strings.Index(s, "-")
			if i < 0 {
				i = len(s)
			}
			j, e = strconv.Atoi(s[:i])
			if e != nil {
				return nil, e
			}
			v.Patch = j

			if i < len(s) {
				v.Label = s[i+1:]
			}
		}
	}

	return &v, nil

}

// Range is a representation of a contiguous set of versions
// The delimiters '[]' and '()' are indicate if the bounds are inclusive or exclusive, respectively
// Example: (1.0.0,2.0.0] represents all versions above 1.0.0 and below or equal to 2.0.0
type Range struct {
	Minimum      Version
	Maximum      Version
	MinInclusive bool
	MaxInclusive bool
}

// String representation of Range
func (r *Range) String() string {
	str := ""
	if r.MinInclusive {
		str += "["
	} else {
		str += "("
	}
	str += r.Minimum.String()
	str += ","
	str += r.Maximum.String()
	if r.MaxInclusive {
		str += "]"
	} else {
		str += ")"
	}
	return str

}

// IsWithin returns true if Version falls within the range specified
func (v Version) IsWithin(r Range) bool {
	switch {
	case r.MinInclusive && r.MaxInclusive:
		return v.Compare(r.Minimum) >= 0 && v.Compare(r.Maximum) <= 0
	case r.MinInclusive && !r.MaxInclusive:
		return v.Compare(r.Minimum) >= 0 && v.Compare(r.Maximum) < 0
	case !r.MinInclusive && r.MaxInclusive:
		return v.Compare(r.Minimum) > 0 && v.Compare(r.Maximum) <= 0
	default: //!d.MinInclusive && !d.MaxInclusive
		return v.Compare(r.Minimum) > 0 && v.Compare(r.Maximum) < 0
	}
}

// ParseRange converts a string representation to a Range.
// Value of 'error' is nil if successful
func ParseRange(s string) (*Range, error) {
	var r Range

	switch s[0] {
	case '[':
		r.MinInclusive = true
	case '(':
		r.MinInclusive = false
	default:
		return nil, errors.New("Range missing [ or (")
	}
	s = s[1:]
	if i := strings.Index(s, ","); i < 0 {
		return nil, errors.New("Range missing ,")
	} else {
		if v, err := ParseVersion(s[:i]); err != nil {
			return nil, err
		} else {
			r.Minimum = *v
		}

		if v, err := ParseVersion(s[i+1 : len(s)-1]); err != nil {
			return nil, err
		} else {
			r.Maximum = *v
		}
		switch s[len(s)-1] {
		case ']':
			r.MaxInclusive = true
		case ')':
			r.MaxInclusive = false
		default:
			return nil, errors.New("Range missing ] or )")
		}

	}

	// Validate range
	switch r.Minimum.Compare(r.Maximum) {
	case 1:
		return nil, errors.New("Minimum version cannot be greater than maximum")
	case 0:
		if !r.MinInclusive || !r.MaxInclusive {
			return nil, errors.New("Single version range should be inclusive")
		}
	}

	return &r, nil

}
