package console

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestLock(c *C) {
	Lock()
	Print("") // Test for deadlocks.
	Unlock()
}
