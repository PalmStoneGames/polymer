package main

import (
	"code.palmstonegames.com/polymer"
	"fmt"
)

func init() {
	polymer.Register(&DataContainer{})
}

type DataContainer struct {
	*polymer.Proto

	AdminData AdminData  `polymer:"bind"`
	UsersData []UserData `polymer:"bind"`
}

type AdminData struct {
	Name        string
	Permissions []string
}

type UserData struct {
	Name string
}

func (t *DataContainer) TagName() string {
	return "data-container"
}

func (t *DataContainer) Created() {
	t.AdminData.Name = "FooBar123"
	t.AdminData.Permissions = []string{"TEST_1", "TEST_2", "TEST_3"}

	t.UsersData = make([]UserData, 5)
	for i := 0; i < 5; i++ {
		t.UsersData[i] = UserData{Name: fmt.Sprintf("USER_TEST_%v", i)}
	}
}

func (t *DataContainer) HandleInput() {
	fmt.Printf("AdminData: %#v\n", t.AdminData)
	fmt.Printf("UsersData: %#v\n", t.UsersData)
}

func main() {}
