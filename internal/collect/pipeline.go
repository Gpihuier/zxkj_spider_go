package collect

type Pipeline interface {
	Process(item any) error
}

type MySQLPipeline struct {
}

func NewMySQLPipeline() Pipeline {
	return &MySQLPipeline{}
}

func (m *MySQLPipeline) Process(item any) error {
	return nil
}
