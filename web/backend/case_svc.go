package main

type ITestCaseService interface {
	Get() (interface{}, error)
	List() (interface{}, error)
	Save() error
}
type CaseSvc struct{}

func (c CaseSvc) Get() (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (c CaseSvc) List() (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (c CaseSvc) Save() error {
	//TODO implement me
	panic("implement me")
}

func NewCaseSvc() ITestCaseService {
	return &CaseSvc{}
}
