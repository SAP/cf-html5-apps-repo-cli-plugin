package models

import (
	"encoding/json"
	"fmt"
)

// HTML5AppDescriptor application descriptor (xs-app.json)
type HTML5AppDescriptor struct {
	AuthenticationMethod *string                   `json:"authenticationMethod,omitempty"`
	Routes               []HTML5AppDescriptorRoute `json:"routes,omitempty"`
}

// HTML5AppDescriptorRoute application descriptor route
type HTML5AppDescriptorRoute struct {
	AuthenticationType *string                       `json:"authenticationType,omitempty"`
	Scope              *HTML5AppDescriptorRouteScope `json:"scope,omitempty"`
}

// HTML5AppDescriptorRouteScope application descriptor route scope
type HTML5AppDescriptorRouteScope struct {
	Verbs   map[string][]string
	Default []string
}

// UnmarshalJSON unmarshal scope
func (s *HTML5AppDescriptorRouteScope) UnmarshalJSON(data []byte) error {
	var scope interface{}
	var err error

	err = json.Unmarshal(data, &scope)
	if err != nil {
		return fmt.Errorf("Failed to parse 'scope' %s: %s", string(data), err.Error())
	}

	switch value := scope.(type) {
	case string:
		s.Default = []string{value}
	case []interface{}:
		s.Default = make([]string, 0)
		for _, item := range value {
			switch itemValue := item.(type) {
			case string:
				s.Default = append(s.Default, itemValue)
			}
		}
	case map[string]interface{}:
		if value["default"] != nil {
			switch defaultValue := value["default"].(type) {
			case string:
				s.Default = []string{defaultValue}
			case []string:
				s.Default = defaultValue
			}
			delete(value, "default")
		}

		for verb := range value {
			switch verbValue := value[verb].(type) {
			case string:
				s.Verbs[verb] = []string{verbValue}
			case []interface{}:
				s.Verbs[verb] = make([]string, 0)
				for _, item := range verbValue {
					switch itemValue := item.(type) {
					case string:
						s.Verbs[verb] = append(s.Verbs[verb], itemValue)
					}
				}
			}
		}
	}
	return nil
}

// GetAllScopes returns set of all scopes of all verbs
func (r *HTML5AppDescriptorRoute) GetAllScopes() []string {
	scopes := make([]string, 0)
	s := r.Scope
	if s != nil {
		for _, verb := range s.Verbs {
		ScopeLoop:
			for _, scope := range verb {
				for _, val := range scopes {
					if val == scope {
						continue ScopeLoop
					}
				}
				scopes = append(scopes, scope)
			}
		}
	DefaultScopeLoop:
		for _, scope := range s.Default {
			for _, val := range scopes {
				if val == scope {
					continue DefaultScopeLoop
				}
			}
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

// GetAllScopes returns set of all scopes of all routes, all verbs
func (d *HTML5AppDescriptor) GetAllScopes() []string {
	scopes := make([]string, 0)
	if d.Routes != nil {
		for _, route := range d.Routes {
			s := route.Scope
			if s != nil {
				for _, verb := range s.Verbs {
				ScopeLoop:
					for _, scope := range verb {
						for _, val := range scopes {
							if val == scope {
								continue ScopeLoop
							}
						}
						scopes = append(scopes, scope)
					}
				}
			DefaultScopeLoop:
				for _, scope := range s.Default {
					for _, val := range scopes {
						if val == scope {
							continue DefaultScopeLoop
						}
					}
					scopes = append(scopes, scope)
				}
			}
		}
	}
	return scopes
}

// IsAuthorizationRequired check if there are routes protected with scopes
func (d *HTML5AppDescriptor) IsAuthorizationRequired() bool {
	if d.AuthenticationMethod == nil || *d.AuthenticationMethod == "xsuaa" {
		return true
	}
	if *d.AuthenticationMethod == "route" {
		if d.Routes == nil || len(d.Routes) == 0 {
			return false
		}
		for _, route := range d.Routes {
			if route.AuthenticationType == nil || *route.AuthenticationType == "xsuaa" {
				return true
			}
		}
	}
	return false
}
