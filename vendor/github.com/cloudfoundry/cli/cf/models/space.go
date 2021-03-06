package models

type SpaceFields struct {
	Guid     string
	Name     string
	AllowSSH bool
}

type Space struct {
	SpaceFields
	Organization     OrganizationFields
	Applications     []ApplicationFields
	ServiceInstances []ServiceInstanceFields
	Domains          []DomainFields
	SecurityGroups   []SecurityGroupFields
	SpaceQuotaGuid   string
}
