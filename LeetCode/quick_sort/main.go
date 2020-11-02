/**
* @name: main
* @author: yuanshuai
* @date: 2020-10-29 11:53
* @description: Description for this page
* @update: 2020-10-29 11:53
 */
package main

import "fmt"

func quickSort(nums []int, start int, end int) {
	if start > end {
		return
	}
	counter := start
	pivot := end

	//pivot作为标杆，counter左边都是小于pivo的值的
	for i := start; i < end; i++ {
		if nums[i] < nums[pivot] {
			nums[counter], nums[i] = nums[i], nums[counter]
			counter++
		}
	}
	nums[counter], nums[pivot] = nums[pivot], nums[counter]

	quickSort(nums, start, counter-1)
	quickSort(nums, counter+1, end)
}

func main() {
	//a := []int{3,2,3,1,2,4,5,5,6}
	a := []int{3, 2, 1, 5, 6, 4}
	quickSort(a, 0, len(a)-1)
	fmt.Println(a)
}
