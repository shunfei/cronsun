package models

// 执行 cron cmd 的进程
// 注册到 /cronsun/proc/<id>
type Node struct {
	ID  string `json:"-"`   // ip
	PID string `json:"pid"` // 进程 pid
}
