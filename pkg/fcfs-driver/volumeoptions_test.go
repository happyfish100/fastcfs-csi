package driver

import (
	"fmt"
	"math"
	"testing"
	"vazmin.github.io/fastcfs-csi/pkg/common"
)

func TestCapacity(t *testing.T) {
	// 1.1g
	gib1piont1 := common.GiB + common.MiB
	fmt.Printf("%f\n", math.Ceil(float64(gib1piont1/common.GiB)))
	fmt.Printf("%f\n", float64(gib1piont1)/float64(common.GiB))
	fmt.Printf("%f\n", math.Ceil(float64(gib1piont1)/float64(common.GiB)))
}
