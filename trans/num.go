package trans

import (
	"reflect"
	"sort"
	"unsafe"
)

func Sort(data []float64) {

	inputSliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&data))

	var dst []int
	dstSliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	*dstSliceHeader = *inputSliceHeader

	sort.Ints(dst)
}
