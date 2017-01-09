package cronsun

type Job struct {
	Id      string   `json:"id"`
	Group   string   `json:"group"`
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Rule    *JobRule `json:"rule"`
}

type JobRule struct {
	Timer        []string `json:"timer"`
	Nodes        []string `json:"nodes"`
	Groups       []string `json:"groups"`
	ExcludeNodes []string `json:"excludeNodes"`
}
