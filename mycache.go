package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
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
	var pid int = os.Getpid()
	fmt.Println("Current process has a PID =", pid)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Current Directory:", cwd)
	}

	var ss []string
	if runtime.GOOS == "windows" {
		ss = strings.Split(cwd, "\\")
	} else {
		ss = strings.Split(cwd, "/")
	}

	currentDirName := ss[len(ss)-1]

	fmt.Println("Current Directory Name:", currentDirName)

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
	}
	myCache := cache.NewLRUCache(10)
	err = myCache.Start(2)
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
	word, _ = myCache.Stats()
	fmt.Println(word)
	_ = myCache.Get("foo1", &val)
	word, _ = myCache.Stats()
	fmt.Println(word)
	_ = myCache.Get("foo2", &val)
	word, _ = myCache.Stats()
	fmt.Println(word)
	_ = myCache.Get("foo3", &val)
	word, _ = myCache.Stats()
	fmt.Println(word)
	_ = myCache.Get("foo1", &val)
	fmt.Println("Print stats")
	word, _ = myCache.Stats()
	fmt.Println(word)
	err = myCache.Purge()
	if err != nil {
		log.Panic(err)
	}
	word, _ = myCache.Stats()
	fmt.Println(word)

	dll := cache.NewDLL(10)
	dll.PrintDLL()
	err = dll.Insert("Hello")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("After hello")
	dll.PrintDLL()
	err = dll.Insert("Boy")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("After boy")
	dll.PrintDLL()
	err = dll.Remove()
	if err != nil {
		log.Panic(err)
	}
	dll.PrintDLL()
	err = dll.Remove()
	if err != nil {
		log.Panic(err)
	}
	dll.PrintDLL()
	myCache.Stop()
}
