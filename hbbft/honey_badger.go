package hbbft

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/sirupsen/logrus"
)

type HoneyBadger struct {
	Config
	// 当前epoch
	epoch uint64
	// epoch -> ACS
	acsInstances map[uint64]*ACS

	// 存事务的buffer
	transactionBuffer *transactionBuffer

	lock sync.Mutex
	// epoch -> []Transaction，在epoch中提交的事务
	outputs map[uint64][]*types.Transaction

	// 需要在处理之后广播的消息
	messageList *messageList

	logger logrus.FieldLogger
}

func NewHoneyBadger(cfg Config) (*HoneyBadger, error) {
	if err := Initialize(); err != nil {
		return nil, err
	}
	return &HoneyBadger{
		Config:            cfg,
		acsInstances:      make(map[uint64]*ACS),
		transactionBuffer: newTransactionBuffer(),
		outputs:           make(map[uint64][]*types.Transaction),
		messageList:       newMessageList(),
		logger:            Logger(honeyBadger),
	}, nil
}

func (hb *HoneyBadger) GetMessage() []Message {
	return hb.messageList.getMessages()
}

// 加到buffer中
func (hb *HoneyBadger) AddTransaction(tx *types.Transaction) {
	hb.transactionBuffer.add(tx)
}

// 处理给定epoch下的ACSMessage
func (hb *HoneyBadger) HandleMessage(sid, epoch uint64, msg *ACSMessage) error {
	acs, ok := hb.acsInstances[epoch]
	if !ok {
		if epoch < hb.epoch {
			return nil
		}
		acs = NewACS(hb.Config)
		hb.acsInstances[epoch] = acs
	}
	if err := acs.HandleMessage(sid, msg); err != nil {
		return err
	}
	hb.addMessages(acs.messageList.getMessages())

	if hb.epoch == epoch {
		acs, ok := hb.acsInstances[hb.epoch]
		if !ok {
			return nil
		}

		outputs := acs.Output()
		if outputs == nil || len(outputs) == 0 {
			return nil
		}

		transactionMap := make(map[string]*types.Transaction)
		for id, output := range outputs {
			hash := CalculateHash(output)
			hb.logger.Debugf("node%d receive commit from node %d in epoch %d, outputHash:%s\n", hb.ID, id, hb.epoch, hash)
			transactions, err := types.UnmarshalTransactions(output)
			if err != nil {
				hb.logger.Errorf("unmarshal transactions error", err)
				return err
			}

			for _, tx := range transactions {
				transactionMap[tx.RbftGetTxHash()] = tx
				hb.logger.Infof("node%d receive commit from node %d in epoch %d, tx[account: %s, nonce:%d]\n", hb.ID, id, hb.epoch, tx.RbftGetFrom(), tx.RbftGetNonce())
			}
		}

		transactions := make([]*types.Transaction, 0)
		for _, tx := range transactionMap {
			transactions = append(transactions, tx)
		}

		hb.transactionBuffer.delete(transactions)
		hb.outputs[hb.epoch] = transactions
		hb.epoch++

		return hb.propose()
	}

	// 移除没用的acs
	for i, acs := range hb.acsInstances {
		if i >= hb.epoch-1 {
			continue
		}
		for _, t := range acs.bbaInstances {
			close(t.closeCh)
		}
		for _, t := range acs.rbcInstances {
			close(t.closeCh)
		}
		close(acs.closeCh)
		delete(hb.acsInstances, i)
	}

	return nil
}

// 调用propose
func (hb *HoneyBadger) Start() error {
	return hb.propose()
}

// 返回每次epoch已经提交的交易
func (hb *HoneyBadger) Outputs() map[uint64][]*types.Transaction {
	hb.lock.Lock()
	defer hb.lock.Unlock()

	out := hb.outputs
	hb.outputs = make(map[uint64][]*types.Transaction)
	return out
}

// 在当前epoch提出一批
func (hb *HoneyBadger) propose() error {
	if hb.transactionBuffer.len() == 0 {
		time.Sleep(2 * time.Second)
	}
	batchSize := hb.BatchSize
	batchSize = int(math.Min(float64(batchSize), float64(hb.transactionBuffer.len())))
	n := int(math.Max(float64(1), float64(batchSize/len(hb.Nodes))))

	rand.Seed(time.Now().UnixNano())
	batch := make([]*types.Transaction, 0)
	for i := 0; i < n; i++ {
		sliceNum := rand.Intn(batchSize)
		// 随机在batchSize中取一笔交易
		batch = append(batch, hb.transactionBuffer.data[:batchSize][sliceNum])
	}
	//batch = append(batch, hb.transactionBuffer.data[:n]...)

	data, err := types.MarshalTransactions(batch)
	if err != nil {
		return err
	}
	hash := CalculateHash(data)
	hb.logger.Debugf("node%d propose %d tx in epoch %d, outputHash:%s\n", hb.ID, len(batch), hb.epoch, hash)

	acs, ok := hb.acsInstances[hb.epoch]
	if !ok {
		acs = NewACS(hb.Config)
	}
	hb.acsInstances[hb.epoch] = acs

	if err := acs.InputValue(data); err != nil {
		return err
	}
	hb.addMessages(acs.messageList.getMessages())
	return nil
}

func (hb *HoneyBadger) addMessages(msgs []Message) {
	for _, msg := range msgs {
		hb.messageList.addMessage(HBMessage{hb.epoch, msg.Payload}, msg.To)
	}
}
