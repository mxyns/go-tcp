package filet

import "strconv"

type Address struct {
	Proto string
	Addr  string
	Port  uint32
}

func (a *Address) ToString() string {

	return a.Addr + ":" + strconv.FormatUint(uint64(a.Port), 10)
}
