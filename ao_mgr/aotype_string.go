// Code generated by "stringer -type AOType"; DO NOT EDIT.

package ao

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeNormal-0]
	_ = x[TypeForced-1]
}

const _AOType_name = "TypeNormalTypeForced"

var _AOType_index = [...]uint8{0, 10, 20}

func (i AOType) String() string {
	if i >= AOType(len(_AOType_index)-1) {
		return "AOType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AOType_name[_AOType_index[i]:_AOType_index[i+1]]
}
