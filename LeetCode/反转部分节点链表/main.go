/**
* @name: main
* @author: yuanshuai
* @date: 2020-10-27 14:58
* @description: Description for this page
* @update: 2020-10-27 14:58
 */
package 反转部分节点链表

type ListNode struct {
	Val  int
	Next *ListNode
}

func reversePartLinkedlist(head *ListNode, m int, n int) *ListNode {
	if head == nil {
		return head
	}

	result := &ListNode{Val: 0, Next: head}
	firstIndex, endIndex := m-1, n-1
	// pre 指向需要反转节点的前一个节点
	pre := result
	for i := 0; i < m-1; i++ {
		pre = pre.Next
	}
	var newHead *ListNode
	cur := pre.Next
	for cur != nil && firstIndex <= endIndex {
		temp := cur.Next
		cur.Next = newHead
		newHead = cur
		cur = temp
		firstIndex++
	}
	pre.Next.Next, pre.Next = cur, newHead
	return result.Next
}
