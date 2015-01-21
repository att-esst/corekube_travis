package util

type HeatConfig struct {
	TemplateFile string
	Keypair      string
	OSAuthUrl    string
	OSUsername   string
	OSPassword   string
	OSTenantId   string
	Timeout      int
}

type HeatStack struct {
	Name            string            `json:"stack_name"`
	Template        string            `json:"template"`
	Params          map[string]string `json:"parameters"`
	Timeout         int               `json:"timeout_mins"`
	DisableRollback bool              `json:"disable_rollback"`
}

type CreateStackResult struct {
	Stack CreateStackResultData `json:"stack"`
}

type CreateStackResultData struct {
	Id    string       `json:"id"`
	Links []StackLinks `json:"links"`
}

type StackLinks struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type StackDetails struct {
	Stack StackDetailsData `json:"stack"`
}

type StackDetailsData struct {
	StackStatus       string                 `json:"stack_status"`
	StackStatusReason string                 `json:"stack_status_reason"`
	Links             []StackLinks           `json:"links"`
	Id                string                 `json:"id"`
	Outputs           []StackDetailsOutput   `json:"outputs"`
	Parameters        map[string]interface{} `json:"parameters"`
}

type StackDetailsOutput struct {
	OutputValue interface{} `json:"output_value"`
	Description string      `json:"description"`
	OutputKey   string      `json:"output_key"`
}
