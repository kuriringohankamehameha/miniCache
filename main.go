package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kuriringohankamehameha/miniCache/cache"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {
	var word string
	if fileExists("test.dump") == true {
		c, err := cache.LoadCache("test.dump")
		if err != nil {
			log.Panic(err)
		}
		var a string
		err = c.Get("fooq", &a)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("'%s'\n", a)
		fmt.Println(c.Stats())
		err = c.Set("foo", "bar", int64(3))
		if err != nil {
			log.Panic(err)
		}
		err = c.Get("foo", &a)
		fmt.Printf("'%s'\n", a)
		if err != nil {
			log.Panic(err)
		}
		word, _ = c.Stats()
		fmt.Println(word)
		os.Exit(0)
	} else {
		myCache := cache.NewLRUCache(10)
		err := myCache.Start(2)
		if err != nil {
			log.Panic(err)
		}
		word, _ = myCache.Stats()
		fmt.Println(word)
		TIMEOUT := 4
		err = myCache.Set("foo", "bar", int64(TIMEOUT))
		if err != nil {
			log.Panic(err)
		}
		word, _ = myCache.Stats()
		fmt.Println(word)
		var val string
		err = myCache.Get("foo", &val)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("'%s'\n", val)
		word, _ = myCache.Stats()
		fmt.Println(word)
		<-time.After(time.Second * time.Duration(TIMEOUT))
		var nilVal string
		err = myCache.Get("foo", &nilVal)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("'%s'\n", nilVal)
		word, _ = myCache.Stats()
		fmt.Println(word)
		for i := 0; i < 5; i++ {
			fmt.Printf("Set i = %d\n", i)
			err = myCache.Set("foo"+strconv.Itoa(i), strings.Repeat("bar", i+1), int64(TIMEOUT))
			if err != nil {
				log.Panic(err)
			}
		}
		err = myCache.Set("fooq", "bbbb", int64(60*TIMEOUT))
		if err != nil {
			log.Panic(err)
		}
		err = myCache.Save("test.dump")
		if err != nil {
			log.Panic(err)
		}
		word, _ = myCache.Stats()
		fmt.Println(word)
		_ = myCache.Get("foo", &val)
		err = myCache.Purge()
		if err != nil {
			log.Panic(err)
		}
		word, _ = myCache.Stats()
		fmt.Println(word)
		myCache.Stop()
	}
}
