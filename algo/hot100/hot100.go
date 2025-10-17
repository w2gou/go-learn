package hot100

type ListNode struct {
	Val  int
	Next *ListNode
}

func hot160_1(listA, listB *ListNode) *ListNode {
	aLast := listA
	bLast := listB
	var aList map[*ListNode]bool
	aList[aLast] = true
	for aLast.Next != nil {
		aLast = aLast.Next
		aList[aLast] = true
	}
	if aList[bLast] {
		return bLast
	}
	for bLast.Next != nil {
		bLast = bLast.Next
		if aList[bLast] {
			return bLast
		}
	}
	return nil
}

func hot160_2(listA, listB *ListNode) *ListNode {
	aLast := listA
	bLast := listB
	for true {
		if aLast == bLast {
			return aLast
		} else {
			if aLast.Next != nil {
				aLast = aLast.Next
			} else {
				aLast = listB
			}
			if bLast.Next != nil {
				bLast = bLast.Next
			} else {
				bLast = listA
			}
		}
	}
	return nil
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func hot236_1(root, p, q *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}

	if root == p || root == q {
		return root
	}

	left := hot236_1(root.Left, p, q)
	right := hot236_1(root.Right, p, q)

	if left != nil && right != nil {
		return root
	}

	if left == nil {
		return right
	}

	return left
}

func hot234_1(heat *ListNode) bool {
	var value []int
	value = append(value, heat.Val)
	for heat.Next != nil {
		heat = heat.Next
		value = append(value, heat.Val)
	}
	leng := len(value)
	if leng%2 != 0 {
		return false
	}
	for i := 0; i < leng/2; i++ {
		if value[i] != value[leng-i-1] {
			return false
		}
	}
	return true
}

func Hot739_1(temperatures []int) []int {
	length := len(temperatures)
	result := make([]int, length)
	var stack []int
	for i := 0; i < length; i++ {
		temperature := temperatures[i]
		for len(stack) > 0 && temperature > temperatures[stack[len(stack)-1]] {
			index := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			result[index] = i - index
		}
		stack = append(stack, i)
	}
	return result
}

func hot226_1(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}
	right := root.Right
	left := root.Left
	root.Left = right
	root.Right = left
	hot226_1(root.Left)
	hot226_1(root.Right)

	return root
}

func hot221_1(matrix [][]byte) int {
	maxSide := 0
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return maxSide
	}

	rows, columns := len(matrix), len(matrix[0])
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			if matrix[i][j] == '1' {
				maxSide = max(maxSide, 1)
				curMaxSide := min(rows-i, columns-j)
				for k := 1; k < curMaxSide; k++ {
					flag := true
					if matrix[i+k][j+k] == '0' {
						break
					}
					for m := 0; m < k; m++ {
						if matrix[i+k][j+m] == '0' {
							flag = false
							break
						}
					}
					if flag {
						maxSide = max(maxSide, k+1)
					} else {
						break
					}
				}
			}
		}
	}
	maxSquare := maxSide * maxSide
	return maxSquare
}

func hot221_2(matrix [][]byte) int {
	maxSide := 0
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return maxSide
	}

	dp := make([][]int, len(matrix))

	rows, columns := len(matrix), len(matrix[0])
	for i := 0; i < rows; i++ {
		dp[i] = make([]int, columns)
		for j := 0; j < columns; j++ {
			dp[i][j] = int(matrix[i][j] - 0)
			if dp[i][j] == 1 {
				maxSide = 1
			}
		}
	}
	for i := 1; i < rows; i++ {
		for j := 1; j < columns; j++ {
			if dp[i][j] == 1 {
				dp[i][j] = min(dp[i-1][j], dp[i-1][j-1], dp[i][j-1]) + 1
				if dp[i][j] > maxSide {
					maxSide = dp[i][j]
				}
			}
		}
	}
	return maxSide * maxSide
}

func hot215_1(nums []int, k int) int {
	n := len(nums)
	for i := 0; i < k; i++ {
		k := i
		for j := i + 1; j < n; j++ {
			if nums[j] > nums[k] {
				k = j
			}
		}
		nums[i], nums[k] = nums[k], nums[i]
	}
	return nums[k-1]
}

func hot215_2(nums []int, k int) int {
	n := len(nums)
	for i := 0; i < k; i++ {
		k := i
		for j := i + 1; j < n; j++ {
			if nums[j] > nums[k] {
				k = j
			}
		}
		nums[i], nums[k] = nums[k], nums[i]
	}
	return nums[k-1]
}

// hot208
type Tire struct {
	children [26]*Tire
	isLeaf   bool
}

func Constructor() Tire {
	return Tire{}
}

func (t *Tire) Insert(word string) {
	node := t
	for _, ch := range word {
		ch -= 'a'
		if node.children[ch] == nil {
			node.children[ch] = &Tire{}
		}
		node = node.children[ch]
	}
}

func (t *Tire) Search(word string) bool {
	node := t
	end := 0
	for _, ch := range word {
		ch -= 'a'
		if node.children[ch] != nil {
			end++
		} else {
			return false
		}
	}
	if end == len(word) {
		return true
	}
	return false
}

func (t *Tire) StartsWith(word string) bool {
	node := t
	end := 0
	for _, ch := range word {
		ch -= 'a'
		if node.children[ch] != nil {
			end++
		} else {
			return false
		}
	}
	return true
}

func hot207_1(numCourses int, prerequisites [][]int) bool {
	var (
		edges   = make([][]int, numCourses)
		visited = make([]int, numCourses)
		result  []int
		valid   = true
		dfs     func(u int)
	)

	// 0-未访问，1-正在访问，2-已完成
	dfs = func(u int) {
		visited[u] = 1
		for _, v := range edges[u] {
			if visited[v] == 0 {
				dfs(v)
				if !valid {
					return
				}
			} else if visited[v] == 1 {
				valid = false
				return
			}
		}
		visited[u] = 2
		result = append(result, u)
	}

	for _, info := range prerequisites {
		edges[info[1]] = append(edges[info[1]], info[0])
	}

	for i := 0; i < numCourses && valid; i++ {
		if visited[i] == 0 {
			dfs(i)
			if !valid {
				break
			}
		}
	}
	return valid
}

func hot207_2(numCourses int, prerequisites [][]int) bool {
	var (
		edges  = make([][]int, numCourses)
		indeg  = make([]int, numCourses)
		result []int
	)

	for _, info := range prerequisites {
		edges[info[1]] = append(edges[info[1]], info[0])
		indeg[info[0]]++
	}

	var q []int
	for i := 0; i < numCourses; i++ {
		if indeg[i] == 0 {
			q = append(q, i)
		}
	}

	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		result = append(result, u)
		for _, v := range edges[u] {
			indeg[v]--
			if indeg[v] == 0 {
				q = append(q, v)
			}
		}
	}
	return len(result) == numCourses
}

func hot206_1(head *ListNode) *ListNode {
	var stack []int
	index := head
	for index.Next != nil {
		stack = append(stack, index.Val)
		index = index.Next
	}
	stack = append(stack, index.Val)

	index = head
	for index.Next != nil {
		index.Val = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		index = index.Next
	}
	index.Val = stack[0]
	return head
}

func hot206_2(head *ListNode) *ListNode {
	var prev *ListNode
	curr := head
	for curr != nil {
		next := curr.Next
		curr.Next = prev
		prev = curr
		curr = next
	}
	return prev
}

func hot200_1(grid [][]byte) int {
	res := 0
	m, n := len(grid), len(grid[0])
	var dfs func(int, int)
	dfs = func(x, y int) {
		if x < 0 || x >= m || y < 0 || y >= n || grid[x][y] != '1' {
			return
		}
		grid[x][y] = '0'
		dfs(x-1, y)
		dfs(x+1, y)
		dfs(x, y-1)
		dfs(x, y+1)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if grid[i][j] == '1' {
				dfs(i, j)
				res++
			}
		}
	}
	return res
}

func hot198_1(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	if len(nums) == 1 {
		return nums[0]
	}
	dp := make([]int, len(nums))
	dp[0] = nums[0]
	dp[1] = max(nums[0], nums[1])
	for i := 2; i < len(nums); i++ {
		dp[i] = max(dp[i-2]+nums[i], dp[i-1])
	}
	return dp[len(nums)-1]
}

func hot169_1(nums []int) int {
	
}
