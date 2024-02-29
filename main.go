package main

import (
	"encoding/json"
	"fmt"
	"memory_cache/cache"
	"time"
)

func main() {
	/**
	使⽤示例
	*/
	sli1 := make([]int8, cache.MB*20)
	bytes1, _ := json.Marshal(sli1) // 50MB

	sli2 := make([]int8, cache.MB*15)
	bytes2, _ := json.Marshal(sli2) // 40mb

	sli3 := make([]int8, cache.MB*15)
	bytes3, _ := json.Marshal(sli3) // 40mb

	sli4 := make([]int8, cache.MB*10)
	bytes4, _ := json.Marshal(sli4) // 20mb

	sli5 := make([]int8, cache.MB*10)
	bytes5, _ := json.Marshal(sli5) // 20mb

	sli6 := make([]int8, cache.MB*10)
	bytes6, _ := json.Marshal(sli6) // 20mb

	var c cache.Cache = cache.NewCache()

	fmt.Println("setMaxMemory = ", c.SetMaxMemory("100mb"))
	fmt.Println("ttl = ", c.Ttl("sli1"))

	_, found := c.Get("sli2")
	if !found {
		fmt.Println("not found")
	}

	fmt.Println("keys = ", c.Keys())
	fmt.Println("del = ", c.Del("sli4"))
	fmt.Println("exists = ", c.Exists("sli1"))

	c.Set("sli1", bytes1)

	// 并发测试
	for i := 0; i < 10; i++ {
		go func() {
			c.Set("sli2", bytes2, 5)
		}()
	}

	fmt.Println("sli1 ttl = ", c.Ttl("sli1"))
	time.Sleep(time.Second)
	fmt.Println("sli2 ttl = ", c.Ttl("sli2"))
	time.Sleep(time.Second * 2)
	c.Set("sli3", bytes3)
	fmt.Println("sli2 exist = ", c.Exists("sli2"))

	c.Set("sli4", bytes4)
	c.Set("sli5", bytes5)
	c.Set("sli6", bytes6)
	fmt.Println("keys = ", c.Keys())
	fmt.Println("sli3 exist = ", c.Exists("sli3"))

	c.Set("sli5", bytes5, 2)
	c.Set("sli3", bytes3)
	time.Sleep(time.Second * 2)
	fmt.Println("sli1 exist = ", c.Exists("sli1"))
	fmt.Println("sli2 exist = ", c.Exists("sli2"))
	fmt.Println("sli3 exist = ", c.Exists("sli3"))
	fmt.Println("sli4 exist = ", c.Exists("sli4"))
	fmt.Println("sli5 exist = ", c.Exists("sli5"))
	fmt.Println("sli6 exist = ", c.Exists("sli6"))

	c.Close()
}
