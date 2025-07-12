package ydn23

import (
	"testing"
)

func TestCalcCHKSUM(t *testing.T) {
	// 示例数据，去除SOI（~）、EOI（CR）和CHKSUM（FD3B）
	data := "20014043E00200"
	want := "FD3B"
	got := CalcCHKSUM(data)
	if got != want {
		t.Errorf("CalcCHKSUM(%q) = %q; want %q", data, got, want)
	}

	// 你可以添加更多测试用例
}
