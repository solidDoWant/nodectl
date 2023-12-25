package node

import "fmt"

// This is keyed by (node number - 1)
// Pulled from Mixtile `nodectl` binary
var serialPortNumbers []uint = []uint{
	1, 2, 3, 0,
}

func (n *Node) BaudRate() uint {
	return 1500000
}

func (n *Node) TTYDevicePath() string {
	return fmt.Sprintf("/dev/ttyCH343USB%d", serialPortNumbers[n.Number()])
}
