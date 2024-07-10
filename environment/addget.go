package environment

//go:generate mockgen -source=$GOFILE -package environment -destination addget_mock.gen.go

// AddGet inteface for representing a data store that can add and get
type AddGet interface {
	// Add adds the passed data to this AddGet
	Add(project string, data map[string]interface{})
	// Get data by project and key for this AddGet
	Get(project, key string) (interface{}, bool)
}

func NewNoopAddGet() AddGet {
	return noop{}
}

type noop struct{}

func (noop) Get(project, key string) (interface{}, bool) { return "", false }

func (noop) Add(project string, data map[string]interface{}) {}
