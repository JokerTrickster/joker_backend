module github.com/JokerTrickster/joker_backend/services/morandoranService

go 1.24.0

toolchain go1.24.1

require (
	github.com/JokerTrickster/joker_backend/shared v0.0.0
	github.com/labstack/echo/v4 v4.13.4
	github.com/stretchr/testify v1.11.1
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.44.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.31.1
)

replace github.com/JokerTrickster/joker_backend/shared => ../../shared
