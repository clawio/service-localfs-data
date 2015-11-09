package lib

import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
}
