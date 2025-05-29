package model

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	s := ServiceTree{
		FullNamePath: "a/b/c",
	}
	path := s.GetSubFullPath()
	fmt.Println(path)
}
