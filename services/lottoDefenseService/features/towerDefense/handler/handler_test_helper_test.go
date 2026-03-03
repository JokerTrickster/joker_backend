package handler

import (
	"encoding/json"
	"testing"

	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const tdTestUserID = uint(1)

func setupTDTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func setupTDAuthContext(c echo.Context) {
	c.Set("userID", tdTestUserID)
}

func tdMustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
