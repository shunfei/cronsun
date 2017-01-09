package models

// 结点类型分组
// 注册到 /cronsun/group/<id>
type Group struct {
	ID   string `json:"-"`
	Name string `json:"name"`

	NodeIDs []string `json:"nids"`
}
