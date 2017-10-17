package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCmdArgumentParser(t *testing.T) {
	var args []string
	var str string

	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 0)
	})

	str = "    "
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 0)
	})

	str = "aa bbb  ccc "
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 3)
		So(args[0], ShouldEqual, "aa")
		So(args[1], ShouldEqual, "bbb")
		So(args[2], ShouldEqual, "ccc")
	})

	str = "' \\\""
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 1)
		So(args[0], ShouldEqual, " \\\"")
	})

	str = `a "b c"` // a "b c"
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 2)
		So(args[0], ShouldEqual, "a")
		So(args[1], ShouldEqual, "b c")
	})

	str = `a '\''"`
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 2)
		So(args[0], ShouldEqual, "a")
		So(args[1], ShouldEqual, "'")
	})

	str = `   \\a   'b c'   c\ d\  `
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 3)
		So(args[0], ShouldEqual, "\\a")
		So(args[1], ShouldEqual, "b c")
		So(args[2], ShouldEqual, "c d ")
	})

	str = `\`
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 1)
		So(args[0], ShouldEqual, "\\")
	})

	str = `  \   ` // \SPACE
	Convey("Parse Cmd Arguments ["+str+"]", t, func() {
		args = ParseCmdArguments(str)
		So(len(args), ShouldEqual, 1)
		So(args[0], ShouldEqual, " ")
	})
}
