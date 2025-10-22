package xlog

import "github.com/labstack/echo/v4"

// Gets the tenant from the context and returns, or attempts to find and set.
// Checks query params for tenantId, header for tenant, param for tenant in that order.
// The first valid found value is returned.
func GetTenant(c echo.Context) string {
	if tenant, ok := c.Get("tenant").(string); ok {
		return tenant
	}

	tenant := c.QueryParam("tenantId")
	if tenant == "" {
		tenant = c.Request().Header.Get("tenant")
		if tenant == "" {
			tenant = c.Param("tenant")
		}
	}

	c.Set("tenant", tenant)
	return tenant
}

// Sets the tenant in the contex and returns
func SetTenant(c echo.Context, tenant string) string {
	c.Set("tenant", tenant)
	return tenant
}
