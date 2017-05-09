package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadExtendConf(t *testing.T) {
	testFile := "test.json"

	type conf struct {
		Debug bool
		Num   int
		Log   struct {
			Level int
			Path  string
		}
	}
	Convey("confutil package test", t, func() {
		Convey("load test file should be success", func() {
			c := &conf{}
			err := LoadExtendConf(testFile, c)
			So(err, ShouldBeNil)
			So(c.Debug, ShouldBeTrue)
			So(c.Log.Path, ShouldEqual, "./tmp")
		})
	})
}
