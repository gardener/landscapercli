package util

import (
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	Testuser = "testuser"
)

var Testpw string

func init() {
	Testpw = string(uuid.NewUUID())
}
