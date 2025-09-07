package repo

import (
	"fmt"
	"strings"
	"testing"
)

func TestName(t *testing.T) {

	s := "/beiluo/demo6/devops/devops_script_create/"
	trim := strings.Trim(s, "/")
	split := strings.Split(trim, "/")
	ss := split[2:]
	fmt.Println(ss)

	sss := split[1:2]
	fmt.Println(sss)
}
