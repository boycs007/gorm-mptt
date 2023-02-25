module github.com/boycs007/gorm-mptt/tests

go 1.16

require (
	github.com/boycs007/gorm-mptt v0.0.0-00010101000000-000000000000
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.24.5
)

replace github.com/boycs007/gorm-mptt => ../
