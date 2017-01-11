package models

// 需要执行的 cron cmd 命令
// 注册到 /cronsun/cmd/<id>
type Job struct {
	ID      string     `json:"-"`
	Name    string     `json:"name"`
	Group   string     `json:"group"`
	Command string     `json:"cmd"`
	Rule    []*JobRule `json:"rule"`
	Status  int        `json:"status"` // 可手工控制的状态，运行中/暂停
}

type JobRule struct {
	Timer          string   `json:"timer"`
	GroupIDs       []string `json:"gids"`
	NodeIDs        []string `json:"nids"`
	ExcludeNodeIDs []string `json:"exclude_nids"`
}
