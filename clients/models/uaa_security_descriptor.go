package models

// UAASecurityDescriptor XSUAA security descriptor (xs-security.json)
type UAASecurityDescriptor struct {
	XSAPPNAME              string                              `json:"xsappname,omitempty"`
	ForeignScopeReferences []string                            `json:"foreign-scope-references,omitempty"`
	Scopes                 []UAASecurityDescriptorScope        `json:"scopes,omitempty"`
	RoleTemplates          []UAASecurityDescriptorRoleTemplate `json:"role-templates,omitempty"`
}

// UAASecurityDescriptorScope XSUAA security descriptor scope
type UAASecurityDescriptorScope struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UAASecurityDescriptorRoleTemplate XSUAA security descriptor role template
type UAASecurityDescriptorRoleTemplate struct {
	Name            *string  `json:"name,omitempty"`
	Description     *string  `json:"description,omitempty"`
	ScopeReferences []string `json:"scope-references,omitempty"`
}
