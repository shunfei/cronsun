package cronsun

type Node struct {
	Pid  int    `json:"pid"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
