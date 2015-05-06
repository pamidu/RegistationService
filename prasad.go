package main

import (
	"duov6.com/objectstore/client"
	"fmt"
)

func main() {
	fmt.Println("1")
	bytes, _ := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByQuerying("EmailAddress :" + "prasadacicts@gmail.com").Ok()

	fmt.Println(string(bytes))
}
