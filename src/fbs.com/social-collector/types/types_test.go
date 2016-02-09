package types

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTypes(t *testing.T) {

	Convey("Social", t, func() {

		Convey("Social empty type is not valid", func() {
			s := Social{}
			So(s.IsValid(), ShouldNotBeNil)
		})
		Convey("Social with UserId = 0 is not valid", func() {
			s := Social{UserId: 0}
			So(s.IsValid(), ShouldNotBeNil)
		})
		Convey("Social empty type with UserId > 0 and not FacebookUrl, TwitterUrl or PhotoUrl is not valid", func() {
			s := Social{UserId: 1}
			So(s.IsValid(), ShouldNotBeNil)
		})

		Convey("Social empty type with UserId > 0 and has FacebookUrl is valid", func() {
			s := Social{UserId: 1, FacebookUrl: "https://test.com"}
			So(s.IsValid(), ShouldBeNil)
		})
		Convey("Social empty type with UserId > 0 and has TwitterUrl is valid", func() {
			s := Social{UserId: 1, TwitterUrl: "https://test.com"}
			So(s.IsValid(), ShouldBeNil)
		})
		Convey("Social empty type with UserId > 0 and has PhotoUrl is valid", func() {
			s := Social{UserId: 1, PhotoUrl: "https://test.com"}
			So(s.IsValid(), ShouldBeNil)
		})

	})
}
