package types

type BaseTime struct {
	/**
	 * 创建时间
	 */
	CreateTime int64 `json:"createTime string"`

	/**
	 * 最后修改时间
	 */
	ModifiedTime int64 `json:"modifiedTime string"`
}

// PluginOut
// swagger:model pluginOut
type PluginOut struct {
	BaseTime
	PluginInfo
	Properties    []Property `json:"properties"`
	IsAnnotateOut bool       `json:"isAnnotateOut"`
	SvcProperties []Property `json:"svcProperties"`
	Out           *Property  `json:"out"`
}

// Property Property
// swagger:model property
type Property struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
