package models

// 需要执行的 cron cmd 命令
// 注册到 /cronsun/cmd/<id>
type Job struct {
	ID      string   `json:"-"`
	Name    string   `json:"name"`
	Group   string   `json:"group"`
	Command string   `json:"cmd"`
	Rule    *JobRule `json:"rule"`
	Status  int      `json:"status"`
}

type JobRule struct {
	Timer          []string `json:"timer"`
	NodeIDs        []string `json:"nids"`
	GroupIDs       []string `json:"gids"`
	ExcludeNodeIDs []string `json:"exclude_bids"`
}
