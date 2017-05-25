package plugin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
	Label string
}

func (v *Version) Compare(other *Version) int {
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

func (v *Version) String() string {
	if v.Label == "" {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Label)

}

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

type Range struct {
	Minimum      Version
	Maximum      Version
	MinInclusive bool
	MaxInclusive bool
}

func (v *Version) IsWithin(r *Range) bool {
	switch {
	case r.MinInclusive && r.MaxInclusive:
		return v.Compare(&r.Minimum) >= 0 && v.Compare(&r.Maximum) <= 0
	case r.MinInclusive && !r.MaxInclusive:
		return v.Compare(&r.Minimum) >= 0 && v.Compare(&r.Maximum) < 0
	case !r.MinInclusive && r.MaxInclusive:
		return v.Compare(&r.Minimum) > 0 && v.Compare(&r.Maximum) <= 0
	default: //!d.MinInclusive && !d.MaxInclusive
		return v.Compare(&r.Minimum) > 0 && v.Compare(&r.Maximum) < 0
	}
}

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

	return &r, nil

}
