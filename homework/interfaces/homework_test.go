package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type UserService struct {
	// not need to implement
	NotEmptyStruct bool
}
type MessageService struct {
	// not need to implement
	NotEmptyStruct bool
}

type CalcService struct {
	User *UserService
}

func NewCalcService(u *UserService) *CalcService {
	return &CalcService{User: u}
}

type Container struct {
	constructors map[string]reflect.Value
	typeToName   map[reflect.Type]string
}

func NewContainer() *Container {
	return &Container{
		constructors: make(map[string]reflect.Value),
		typeToName:   make(map[reflect.Type]string),
	}
}

func (c *Container) RegisterType(name string, constructor interface{}) {
	v := reflect.ValueOf(constructor)

	if v.Kind() != reflect.Func {
		panic("constructor must be a function")
	}

	if _, exists := c.constructors[name]; exists {
		return
	}

	ct := v.Type()
	if ct.NumOut() == 0 {
		panic("constructor must return the value")
	}

	c.constructors[name] = v

	returnType := ct.Out(0)

	c.typeToName[returnType] = name
}

func (c *Container) Resolve(name string) (interface{}, error) {
	found := make(map[string]bool)

	v, err := c.resolve(name, found)
	if err != nil {
		return nil, err
	}
	return v.Interface(), nil
}

func (c *Container) resolve(name string, found map[string]bool) (reflect.Value, error) {
	if found[name] {
		return reflect.Value{}, fmt.Errorf("cyclical relationship has been discovered %q", name)
	}

	cons, ok := c.constructors[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("no constructor registered for %q", name)
	}

	found[name] = true
	defer delete(found, name)

	consTypes := cons.Type()
	args := make([]reflect.Value, 0, consTypes.NumIn())

	for i := 0; i < consTypes.NumIn(); i++ {
		paramType := consTypes.In(i)

		srvName, ok := c.typeToName[paramType]
		if !ok {
			return reflect.Value{}, fmt.Errorf("there is no registration for the dependency type %s", paramType)
		}

		srvValue, err := c.resolve(srvName, found)
		if err != nil {
			return reflect.Value{}, err
		}

		args = append(args, srvValue)
	}

	results := cons.Call(args)

	if len(results) == 2 {
		if errVal := results[1]; !errVal.IsNil() {
			return reflect.Value{}, errVal.Interface().(error)
		}
	}

	return results[0], nil
}

func TestDIContainer(t *testing.T) {
	container := NewContainer()
	container.RegisterType("UserService", func() *UserService {
		return &UserService{NotEmptyStruct: true}
	})
	container.RegisterType("MessageService", func() interface{} {
		return &MessageService{}
	})

	userService1, err := container.Resolve("UserService")
	assert.NoError(t, err)
	userService2, err := container.Resolve("UserService")
	assert.NoError(t, err)

	u1 := userService1.(*UserService)
	u2 := userService2.(*UserService)
	assert.False(t, u1 == u2)

	messageService, err := container.Resolve("MessageService")
	assert.NoError(t, err)
	assert.NotNil(t, messageService)

	paymentService, err := container.Resolve("PaymentService")
	assert.Error(t, err)
	assert.Nil(t, paymentService)

	container.RegisterType("CalcService", NewCalcService)
	calcService, err := container.Resolve("CalcService")
	assert.NoError(t, err)
	assert.NotNil(t, messageService)

	assert.NoError(t, err)
	calc := calcService.(*CalcService)

	assert.NotNil(t, calc.User)
	assert.True(t, calc.User.NotEmptyStruct)
}
