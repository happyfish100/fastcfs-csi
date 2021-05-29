package fcfs

import (
	"fmt"
	"github.com/happyfish100/fastcfs-csi/pkg/common"
	"math"
	"testing"
)

func TestCapacity(t *testing.T) {
	// 1.1g
	gib1piont1 := common.GiB + common.MiB
	fmt.Printf("%f\n", math.Ceil(float64(gib1piont1/common.GiB)))
	fmt.Printf("%f\n", float64(gib1piont1)/float64(common.GiB))
	fmt.Printf("%f\n", math.Ceil(float64(gib1piont1)/float64(common.GiB)))
}
