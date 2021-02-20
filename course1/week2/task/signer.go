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

	lock := &sync.Mutex{}

	wg := &sync.WaitGroup{}

	for data := range in {

		wg.Add(1)
		dataStr := strconv.Itoa(data.(int))
		go gatherSingleHash(dataStr, wg, lock, out)
	}

	wg.Wait()
}

func gatherSingleHash(data string, dataWg *sync.WaitGroup, lock *sync.Mutex, out chan interface{}) {

	defer dataWg.Done()

	lock.Lock()
	data_md5 := DataSignerMd5(data)
	lock.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	rez := make([]string, 2)

	go chanDataSignerCrc32(0, data, rez, wg)
	go chanDataSignerCrc32(1, data_md5, rez, wg)

	wg.Wait()

	out <- strings.Join(rez, "~")

}

func chanDataSignerCrc32(n int, data string, rez []string, wg *sync.WaitGroup) {
	defer wg.Done()

	rez[n] = DataSignerCrc32(data)

}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
// где th=0..5 ( т.е. 6 хешей на каждое входящее значение ), потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {

	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		dataStr := data.(string)

		go gatherMultiHash(dataStr, wg, out)

	}

	wg.Wait()
}

func gatherMultiHash(data string, dataWg *sync.WaitGroup, out chan interface{}) {

	defer dataWg.Done()

	const calls = 6

	wg := &sync.WaitGroup{}
	rez := make([]string, 6)
	wg.Add(calls)

	for th := 0; th < calls; th++ {

		go chanDataSignerCrc32(th, strconv.Itoa(th)+data, rez, wg)

	}

	wg.Wait()

	out <- strings.Join(rez, "")

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
