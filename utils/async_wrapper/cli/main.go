package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type InputMsg struct {
	input string
	err   error
}

func InputDispathcer() chan *InputMsg {
	inputChan := make(chan *InputMsg)
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "err") {
				inputChan <- &InputMsg{
					"", errors.New(line),
				}
			} else {
				inputChan <- &InputMsg{
					line, nil,
				}
			}
		}
	}()
	return inputChan
}

func GetInput(c chan *InputMsg) *async_wrapper.AsyncWrapper[string] {
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[string]) {
		select {
		case <-ac.Context().Done():
			// async call canceled
			break
		case msg := <-c:
			if msg.err != nil {
				ac.SetErrIfNo(msg.err)
			} else {
				ac.SetResult(msg.input)
			}
		}
	}, true)
}

func main() {
	inputChan := InputDispathcer()
	// fmt.Println(GetInput(inputChan).SetTimeout(time.Second * 3).BlockGetResult())
	GetInput(inputChan).AsyncGetResult(func(ret string, err error) {
		fmt.Println(ret, err)
	})
	// GetInput(inputChan).AsyncOmitResult()
	// GetInput(inputChan).RedirectResult()
	fmt.Println("finish")
	time.Sleep(time.Hour)
}
