package trans

import (
	"fmt"
	"testing"
)

func TestSort(t *testing.T) {
	s := []float64{1.1, 2.2, 0.5, 3.2, 1.2, 3, 2, 1}
	Sort(s)

	for _, v := range s {
		fmt.Println(v)
	}
}
