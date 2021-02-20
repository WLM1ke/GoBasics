package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// ExecutePipeline of jobs
func ExecutePipeline(jobs ...job) {

	wg := &sync.WaitGroup{}

	chans := createArrayOfChan(len(jobs))
	close(chans[0])

	for n, step := range jobs {

		wg.Add(1)
		go jobStarter(step, chans[n], chans[n+1], wg)

	}
	wg.Wait()

}

func createArrayOfChan(nJobs int) []chan interface{} {

	chans := make([]chan interface{}, nJobs+1)

	for i := range chans {
		chans[i] = make(chan interface{})
	}

	return chans
}

func jobStarter(step job, in, out chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()
	defer close(out)

	step(in, out)

}

// SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~),
// где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {

	for data := range in {

		dataStr := strconv.Itoa(data.(int))
		out <- DataSignerCrc32(dataStr) + "~" + DataSignerCrc32(DataSignerMd5(dataStr))
	}
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
// где th=0..5 ( т.е. 6 хешей на каждое входящее значение ), потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {

	for data := range in {

		dataStr := data.(string)

		var outStr string

		for th := 0; th <= 5; th++ {

			outStr += DataSignerCrc32(strconv.Itoa(th) + dataStr)

		}
		out <- outStr
	}
}

// CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/),
// объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {

	rez := []string{}

	for data := range in {

		rez = append(rez, data.(string))

	}

	sort.Strings(rez)

	out <- strings.Join(rez, "_")
}
