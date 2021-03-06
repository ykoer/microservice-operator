package controller

import (
	"github.com/ykoer/microservice-operator/pkg/controller/microservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, microservice.Add)
}
