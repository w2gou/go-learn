package sort

import (
	"math/rand"
	"time"
)

// 选择排序
func selectionSort(nums []int) {
	n := len(nums)
	for i := 0; i < n-1; i++ {
		k := i
		for j := i + 1; j < n; j++ {
			if nums[j] < nums[k] {
				k = j
			}
		}
		nums[k], nums[i] = nums[i], nums[k]
	}
}

// 冒泡排序
func bubbleSort(nums []int) {
	n := len(nums)
	for i := n - 1; i > 0; i-- {
		for j := 0; j < i; j++ {
			if nums[j] > nums[j+1] {
				nums[j], nums[j+1] = nums[j+1], nums[j]
			}
		}
	}
}

// 冒泡排序-标志优化
func bubbleSortWithFlag(nums []int) {
	n := len(nums)
	for i := n - 1; i > 0; i-- {
		flag := false
		for j := 0; j < i; j++ {
			if nums[j] > nums[j+1] {
				nums[j], nums[j+1] = nums[j+1], nums[j]
				flag = true
			}
		}
		if flag == false {
			break
		}
	}
}

// 插入排序
func insertionSort(nums []int) {
	n := len(nums)
	for i := 1; i < n; i++ {
		base := nums[i]
		j := i - 1
		for j >= 0 && nums[j] > base {
			nums[j+1] = nums[j]
			j--
		}
		nums[j+1] = base
	}
}

// 希尔排序-基于插入排序
func shellSort(nums []int) {
	length := len(nums)
	gap := 1
	for gap < length/3 {
		gap = gap*3 + 1
	}
	for gap > 0 {
		for i := gap; i < length; i++ {
			tem := nums[i]
			j := i - gap
			for j > 0 && nums[j] > tem {
				nums[j+gap] = nums[j]
				j -= gap
			}
			nums[j+gap] = tem
		}
		gap = gap / 3
	}
}

// 归并排序
func MergeSort(nums []int) {
	if len(nums) < 2 {
		return
	}
	buf := make([]int, len(nums))
	mergeSort(nums, buf, 0, len(nums))
}

func mergeSort(nums, buf []int, left, right int) {
	if right-left <= 1 {
		return
	}
	mid := (left + right) / 2
	mergeSort(nums, buf, left, mid)
	mergeSort(nums, buf, mid, right)
	merge(nums, buf, left, mid, right)
}

func merge(nums, buf []int, left, mid, right int) {
	i, j, k := left, mid, left
	for i < mid && j < right {
		if nums[i] < nums[j] {
			buf[k] = nums[i]
			i++
		} else {
			buf[k] = nums[j]
			j++
		}
		k++
	}
	for i < mid {
		buf[k] = nums[i]
		i++
		k++
	}
	for j < right {
		buf[k] = nums[j]
		j++
		k++
	}
	copy(nums[left:right], buf[left:right])
}

// 快速排序
func QuickSort(nums []int) {
	if len(nums) < 2 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	quickSort(nums, 0, len(nums)-1)
}

func quickSort(nums []int, left, right int) {
	for left < right {
		pivotIdx := left + rand.Intn(right-left+1)
		nums[left], nums[pivotIdx] = nums[pivotIdx], nums[left]
		p := partition(nums, left, right)

		if p-left < right-p {
			quickSort(nums, left, p-1)
			left = p + 1
		} else {
			quickSort(nums, p+1, right)
			right = p - 1
		}
	}
}

func partition(nums []int, left, right int) int {
	pivot := nums[left]
	i := left + 1
	for j := left + 1; j <= right; j++ {
		if nums[j] < pivot {
			nums[j], nums[i] = nums[i], nums[j]
			i++
		}
	}
	nums[left], nums[i-1] = nums[i-1], nums[left]
	return i - 1
}

// 堆排序
func HeapSort(nums []int) {
	n := len(nums)
	if n < 2 {
		return
	}
	// 1) 建堆：从最后一个非叶子节点开始下沉
	for i := n/2 - 1; i >= 0; i-- {
		siftDown(nums, i, n)
	}
	// 2) 反复把堆顶（最大值）交换到末尾，并缩小堆
	for end := n - 1; end > 0; end-- {
		nums[0], nums[end] = nums[end], nums[0]
		siftDown(nums, 0, end)
	}
}

// 大根堆
func siftDown(nums []int, i, n int) {
	for {
		left := i*2 + 1
		right := left + 1
		largest := i
		if left < n && nums[left] > nums[largest] {
			largest = left
		}
		if right < n && nums[right] > nums[largest] {
			largest = right
		}
		if largest == i {
			return
		}
		nums[i], nums[largest] = nums[largest], nums[i]
		i = largest
	}
}
