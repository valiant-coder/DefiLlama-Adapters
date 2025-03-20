package admin

import (
	"context"
	"fmt"
	"exapp-go/api"
	"exapp-go/internal/db/db"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-gonic/gin"
)

type CasbinAdapter struct {
}

func (c *CasbinAdapter) LoadPolicy(model model.Model) error {

	persist.LoadPolicyLine("p, ROLE_ADMIN, *, *", model)

	repo := db.New()
	ctx := context.Background()
	roles, err := repo.GetAllAdminRoles(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, permission := range role.Permissions {
			for _, action := range permission.Actions {
				persist.LoadPolicyLine(fmt.Sprintf("p, ROLE_%s, %s, %s", strings.ToUpper(role.Name), action.Uri, action.Method), model)
			}
		}
	}

	var admins []*db.Admin
	admins, err = repo.GetAllAdmins(ctx)
	if err != nil {
		return err
	}
	for _, admin := range admins {
		for _, role := range admin.Roles {
			persist.LoadPolicyLine(fmt.Sprintf("g, %s, ROLE_%s", admin.Name, strings.ToUpper(role.Name)), model)
		}
	}

	return nil
}

func (c *CasbinAdapter) SavePolicy(model model.Model) error {
	return nil
}

func (c *CasbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (c *CasbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (c *CasbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func Authorized(enforcer *casbin.SyncedEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := c.GetString("admin")
		method := c.Request.Method
		uri := c.Request.URL.Path

		ok,err := enforcer.Enforce(admin, uri, method)
		if err != nil {
			api.Error(c, err)
			return
		}
		if !ok {
			api.NoPermission(c)
			return
		}

		c.Next()
	}
}
