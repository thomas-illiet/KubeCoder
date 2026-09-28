package organizations

import "testing"

// TestOrganizationOrderClause verifies all supported orders and safe fallbacks.
func TestOrganizationOrderClause(t *testing.T) {
	tests := []struct {
		name      string
		orderBy   OrganizationOrder
		direction OrderDirection
		want      string
	}{
		{name: "name ascending", orderBy: OrganizationOrderName, direction: OrderAscending, want: "organizations.name ASC, organizations.id ASC"},
		{name: "creation descending", orderBy: OrganizationOrderCreatedAt, direction: OrderDescending, want: "organizations.created_at DESC, organizations.id ASC"},
		{name: "members ascending", orderBy: OrganizationOrderMemberCount, direction: OrderAscending, want: "member_count ASC, organizations.id ASC"},
		{name: "repositories descending", orderBy: OrganizationOrderRepositoryCount, direction: OrderDescending, want: "repository_count DESC, organizations.id ASC"},
		{name: "unknown values are safe defaults", orderBy: OrganizationOrder("unsafe"), direction: OrderDirection("unsafe"), want: "organizations.name ASC, organizations.id ASC"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := organizationOrderClause(test.orderBy, test.direction); got != test.want {
				t.Fatalf("organizationOrderClause() = %q, want %q", got, test.want)
			}
		})
	}
}

// TestMemberOrderClause verifies all supported member orders and safe fallbacks.
func TestMemberOrderClause(t *testing.T) {
	tests := []struct {
		name      string
		orderBy   MemberOrder
		direction OrderDirection
		want      string
	}{
		{name: "display name ascending", orderBy: MemberOrderDisplayName, direction: OrderAscending, want: "users.display_name ASC, users.id ASC"},
		{name: "username descending", orderBy: MemberOrderUsername, direction: OrderDescending, want: "users.username DESC, users.id ASC"},
		{name: "email ascending", orderBy: MemberOrderEmail, direction: OrderAscending, want: "users.email ASC, users.id ASC"},
		{name: "joined descending", orderBy: MemberOrderJoinedAt, direction: OrderDescending, want: "organization_memberships.created_at DESC, users.id ASC"},
		{name: "unknown values are safe defaults", orderBy: MemberOrder("unsafe"), direction: OrderDirection("unsafe"), want: "users.display_name ASC, users.id ASC"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := memberOrderClause(test.orderBy, test.direction); got != test.want {
				t.Fatalf("memberOrderClause() = %q, want %q", got, test.want)
			}
		})
	}
}
