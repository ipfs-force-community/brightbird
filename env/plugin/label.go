package plugin

import (
	"github.com/hunjixin/brightbird/types"
)

const SvcName = "svcname"

const ignoreByFront = "ignore"

const CodeVersionPropName = "codeVersion"

func FindCodeVersionProperties(properties []*types.Property) *types.Property {
	for _, property := range properties {
		if property.Name == CodeVersionPropName {
			return property
		}
	}
	return nil
}
