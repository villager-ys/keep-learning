/**
* @name: twoSum
* @author: yuanshuai
* @date: 2020-09-16 21:30
* @description: Description for this page
* @update: 2020-09-16 21:30
 */
package main

type ListNode struct {
	Val  int
	Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	l := new(ListNode)
	result := l
	tmp := 0
	for l1 != nil || l2 != nil || tmp != 0 {
		if l1 != nil {
			tmp += l1.Val
			l1 = l1.Next
		}
		if l2 != nil {
			tmp += l2.Val
			l2 = l2.Next
		}
		l.Next = &ListNode{tmp % 10, nil}
		tmp = tmp / 10
		l = l.Next
	}
	return result.Next
}

func main() {
	l1 := ListNode{
		Val: 5,
		Next: &ListNode{
			Val:  4,
			Next: nil,
		},
	}
	l2 := ListNode{
		Val:  6,
		Next: nil,
	}
	l := addTwoNumbers(&l1, &l2)
	print(l)
}
