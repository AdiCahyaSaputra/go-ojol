package script

import "fmt"

type ExampleScript struct{}

func NewExampleScript() *ExampleScript {
	return &ExampleScript{}
}

func (s *ExampleScript) Run() error {
	fmt.Println("example script running")
	return nil
}
