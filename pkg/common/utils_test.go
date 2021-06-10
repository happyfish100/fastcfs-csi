package common

import "testing"

func TestRoundUpGiB(t *testing.T) {
	var sizeInBytes int64 = 1073741824
	actual := RoundUpGiB(sizeInBytes)
	if actual != 1 {
		t.Fatalf("Wrong result for RoundUpGiB. Got: %d", actual)
	}
}

func TestGiBToBytes(t *testing.T) {
	actual := GiBToBytes(1)
	if actual != GiB {
		t.Fatalf("Wrong result for RoundUpGiB. Got: %d", actual)
	}
}

func TestRoundOffBytes(t *testing.T) {
	actual := RoundOffBytes(GiB)
	if actual != GiB {
		t.Fatalf("Wrong result for RoundOffBytes. Got: %d", actual)
	}
}
