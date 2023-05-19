package plugin

import (
	"github.com/hunjixin/brightbird/types"
)

const InstancePropertyName = "instanceName"

const Optional = "optional"

const SvcName = "svcname"

const CodeVersionPropName = "codeVersion"

func FindCodeVersionProperties(properties []*types.Property) *types.Property {
	for _, property := range properties {
		if property.Name == CodeVersionPropName {
			return property
		}
	}
	return nil
}
