package role

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestRolesCan(t *testing.T) {
	t.Parallel()

	projectID := uuid.NewV4().String()

	employeeRole := Role{
		Name:    "employee",
		Label:   "Employee",
		Implies: Roles{},
	}

	allTeamLeadRole := Role{
		Name:    "teamlead",
		Label:   "Team Lead",
		Over:    []string{"*"},
		Implies: Roles{employeeRole},
	}

	teamLeadRole := Role{
		Name:    "teamlead",
		Label:   "Team Lead",
		Over:    []string{projectID},
		Implies: Roles{employeeRole},
	}

	adminRole := Role{
		Name:    "admin",
		Label:   "Admin",
		Implies: Roles{allTeamLeadRole},
	}

	admin := Roles{}
	admin = append(admin, adminRole)

	teamleader := Roles{}
	teamleader = append(teamleader, teamLeadRole)

	phantomProjectID := uuid.NewV4().String()

	assert.True(t, admin.CanOver(teamLeadRole.Name, phantomProjectID))
	assert.True(t, admin.Can(teamLeadRole.Name))
	assert.True(t, admin.Can(employeeRole.Name))
	assert.False(t, teamleader.CanOver(teamLeadRole.Name, phantomProjectID))
	assert.True(t, teamleader.Can(teamLeadRole.Name))
	assert.True(t, teamleader.Can(employeeRole.Name))
	assert.True(t, teamleader.CanOver(teamLeadRole.Name, projectID))
}
