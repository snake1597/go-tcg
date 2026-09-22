package game

// Input 是 controller 提交給單局引擎的一次玩家操作。
// Action、Choice 與付款欄位只接受目前 PlayerView 所提供的 handle。
type Input struct {
	Revision uint64     `json:"revision"`
	Action   ViewHandle `json:"action"`
	Choice   ViewHandle `json:"choice"`
	// Reserve 只接受目前玩家可見的手牌 handles，並由 activation 的 ReserveCost 決定精確張數。
	Reserve []ViewHandle `json:"reserve,omitempty"`
	// MemoryPayment 只接受當前 action 提供的非隨機 Memory 付款來源。
	MemoryPayment []ViewHandle `json:"memory_payment,omitempty"`
}
