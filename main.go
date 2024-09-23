package main

import (
	"fmt"
	"sync"
)

// Число сообщений от источника
const messagesAmountPerGoroutine int = 5

// Функция разуплотнения каналов
func demultiplexingFunc(dataSourceChan chan int, amount int) []chan int {
	var output = make([]chan int, amount)
	for i := range output {
		output[i] = make(chan int)
	}

	done := make(chan struct{})

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range dataSourceChan {
				for _, c := range output {
					c <- v
				}
			}
			close(done)
		}()
		wg.Wait()
	}()

	go func() {
		<-done
		for _, c := range output {
			close(c)
		}
	}()

	return output
}

// Функция уплотнения каналов
func multiplexingFunc(channels ...chan int) <-chan int {
	var wg sync.WaitGroup
	multiplexedChan := make(chan int)
	multiplex := func(c <-chan int) {
		defer wg.Done()
		for i := range c {
			multiplexedChan <- i
		}
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}
	go func() {
		wg.Wait()
		close(multiplexedChan)
	}()
	return multiplexedChan
}

func main() {
	startDataSource := func() chan int {
		c := make(chan int)
		go func() {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 1; i <= messagesAmountPerGoroutine; i++ {
					c <- i
				}
			}()
			wg.Wait()
			close(c)
		}()
		return c
	}

	consumers := demultiplexingFunc(startDataSource(), 5)
	c := multiplexingFunc(consumers...)

	for data := range c {
		fmt.Println(data)
	}
	fmt.Print('1')
	fmt.Print('1')
	fmt.Print('1')
	fmt.Print('1')
}
