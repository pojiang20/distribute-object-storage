package rabbitmq

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	channel      *amqp.Channel
	Name         string
	exchangeName string
}

func New(s string) *RabbitMQ {
	conn, err := amqp.Dial(s)
	err_utils.Panic_NonNilErr(err)

	ch, err := conn.Channel()
	err_utils.Panic_NonNilErr(err)

	q, err := ch.QueueDeclare("", false, true, false, false, nil)
	err_utils.Panic_NonNilErr(err)

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.Name = q.Name
	return mq
}

// 当前队列与exchange绑定
func (q *RabbitMQ) BindExchange(exchangeName string) {
	err := q.channel.QueueBind(q.Name, "", exchangeName, false, nil)
	err_utils.Panic_NonNilErr(err)
	q.exchangeName = exchangeName
}

// 用于向特定队列发送消息
func (q *RabbitMQ) Send(queueName string, msgBody interface{}) {
	str, err := json.Marshal(msgBody)
	err_utils.Panic_NonNilErr(err)
	err = q.channel.Publish("", queueName, false, false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	err_utils.Panic_NonNilErr(err)
}

// 消息发送给所有绑定了该exchange的队列
func (q *RabbitMQ) Publish(exchangeName string, msgBody interface{}) {
	str, err := json.Marshal(msgBody)
	err_utils.Panic_NonNilErr(err)
	err = q.channel.Publish(exchangeName, "", false, false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	err_utils.Panic_NonNilErr(err)
}

func (q *RabbitMQ) Consume() <-chan amqp.Delivery {
	c, err := q.channel.Consume(q.Name, "", true, false, false, false, nil)
	err_utils.Panic_NonNilErr(err)
	return c
}

func (q *RabbitMQ) Close() {
	q.channel.Close()
}
