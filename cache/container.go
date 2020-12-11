package cache

import (
	"errors"
    "fmt"
)

type DLL struct {
    /*
    Doubly Linked list Structure:
    Container for the LRU Cache
     */
    list *DLLNode
    Size int
    Capacity int
}

type DLLNode struct {
    /*
    A Doubly Linked list node for the linkages in the eviction table
     */
    value string
    next *DLLNode
    prev *DLLNode
}

func NewDLL(capacity int) *DLL {
    // Constructs a new Doubly Linked list
    dll := &DLL{
        list: nil,
        Size: 0,
        Capacity: capacity,
    }
    return dll
}

func (dll* DLL) Insert(value string) error {
    return dll.insertHead(value)
}

func (dll* DLL) Remove() error {
    return dll.removeTail()
}

func (dll* DLL) PrintDLL() {
    head := dll.list
    if head == nil {
        fmt.Println("list is Empty!")
    } else {
        curr := 0
        for temp:=head; ; temp=temp.next {
            fmt.Printf("=> '%s' ", temp.value)
            if curr + 1 >= dll.Size {
                break
            }
            curr++
        }
        fmt.Printf("\n")
    }
}

func (dll* DLL) clear() error {
    // Resets the state of the DLL
    dll.Size = 0
    dll.list = nil
    return nil
}

func (dll *DLL) insertHead(value string) error {
    // Inserts the value at the head of the DLL
    if dll.Size >= dll.Capacity {
        return errors.New("buffer capacity is full")
    }

    node := &DLLNode{
        value: value,
        next: nil,
        prev: nil,
    }

    if dll.list == nil {
        dll.list = node
        node.prev = node
        node.next = node
        dll.Size++
    } else {
        head := dll.list
        prevNode := head.prev
        prevNode.next = node
        head.prev = node
        node.prev = prevNode
        node.next = head
        dll.list = node
        dll.Size++
    }
    return nil
}

func (dll *DLL) shiftToHead(node *DLLNode, insertNode bool) error {
    /*
    Shifts a node to the head position of the DLL
    If this node exists already, then insertNode must be set to false
    */
    if insertNode == true {
        if dll.Size < dll.Capacity {
        } else {
            err := dll.removeTail()
            if err != nil {
                return err
            }
        }
    }
    if dll.list == nil {
        node.prev = node
        node.next = node
        dll.list = node
    } else {
        head := dll.list
        nodePrev := node.prev
        nodeNext := node.next
        if nodePrev != nil && nodeNext != nil {
            nodePrev.next = nodeNext
            nodeNext.prev = nodePrev
        }
        prevNode := head.prev
        prevNode.next = node
        head.prev = node
        node.prev = prevNode
        node.next = head
        dll.list = node
    }
    if insertNode == true {
        dll.Size++
    }
    return nil
}

func (dll* DLL) removeNode(node *DLLNode) error {
    if dll.list == nil {
    } else {
        prevNode := node.prev
        nextNode := node.next
        if prevNode != nil {
            prevNode.next = nextNode
        }
        if nextNode != nil {
            nextNode.prev = prevNode
        }
        node = nil
    }
    return nil
}

func (dll *DLL) removeTail() error {
    /*
    Removes a node from the tail of the DLL
     */
    if dll.Size <= 1 {
        dll.list = nil
        dll.Size = 0
        return nil
    } else {
        head := dll.list
        tail := dll.list.prev
        prevNode := tail.prev
        head.prev = prevNode
        prevNode.next = head
        tail = nil
        dll.list = head
        dll.Size--
        return nil
    }
}
