package rocketmq

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/Hyingerrr/mirco-esim/core/tracer"

	"github.com/opentracing/opentracing-go"

	"github.com/Hyingerrr/mirco-esim/config"
	"github.com/Hyingerrr/mirco-esim/log"

	mq_http_sdk "github.com/aliyunmq/mq-http-go-sdk"
)

type Publisher struct {
	client *MQClient

	logger log.Logger

	conf config.Config

	allowTracer bool
}

type PublisherOption func(*Publisher)

func NewPublisher(options ...PublisherOption) *Publisher {

	p := &Publisher{}

	for _, option := range options {
		option(p)
	}

	if p.conf == nil {
		p.conf = config.NewNullConfig()
	}

	if p.logger == nil {
		p.logger = log.NewLogger()
	}

	p.allowTracer = p.conf.GetBool("mq_publisher_trace")

	/*初始化客户端*/
	var cliOpt MQClientOptions
	p.client = NewMQClient(
		cliOpt.WithLogger(p.logger),
		cliOpt.WithConf(p.conf))
	return p
}

func WithPublisherConf(conf config.Config) PublisherOption {
	return func(p *Publisher) {
		p.conf = conf
	}
}

func WithPublisherLogger(logger log.Logger) PublisherOption {
	return func(p *Publisher) {
		p.logger = logger
	}
}

func (p *Publisher) PublishMessage(ctx context.Context, topicName, messageBody string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
	}
	return p.publish(ctx, topicName, msg)
}

// with message tag and key
func (p *Publisher) PublishMsgWithTag(ctx context.Context, topicName, messageBody, messageTag string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
		MessageTag:  messageTag,
	}
	return p.publish(ctx, topicName, msg)
}

// with message tag
func (p *Publisher) PublishMsgWithKeyTag(ctx context.Context, topicName, messageBody, messageTag, messageKey string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
		MessageTag:  messageTag,
		MessageKey:  messageKey,
	}
	return p.publish(ctx, topicName, msg)
}

func (p *Publisher) PublishDelayMessage(ctx context.Context, topicName, messageBody string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(ctx, topicName, msg)
}

func (p *Publisher) PublishDelayMsgWithTag(ctx context.Context, topicName, messageBody, messageTag string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		MessageTag:       messageTag,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(ctx, topicName, msg)
}

func (p *Publisher) PublishDelayMsgWithKeyTag(ctx context.Context, topicName, messageBody, messageTag, messageKey string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		MessageTag:       messageTag,
		MessageKey:       messageKey,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(ctx, topicName, msg)
}

func (p *Publisher) PublishMessageProp(ctx context.Context, topicName, messageTag, messageBody string, properties map[string]string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
		MessageTag:  messageTag, // 消息标签
		Properties:  properties,
	}
	return p.publish(ctx, topicName, msg)
}

func (p *Publisher) PublishDelayMessageProt(ctx context.Context, topicName, messageBody string, properties map[string]string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		Properties:       properties,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(ctx, topicName, msg)
}

func (p *Publisher) publish(ctx context.Context, topicName string, msg mq_http_sdk.PublishMessageRequest) (err error) {
	var resp mq_http_sdk.PublishMessageResponse
	{
		if !p.allowTracer {
			goto Next
		}

		tc := p.withTrace(ctx)
		childSpan := tc.tracer.StartSpan(
			"RocketMQ_publisher",
			opentracing.ChildOf(tc.spanCtx),
		)

		defer func() {
			if err != nil {
				ext.Error.Set(childSpan, true)
				ext.MessageBusDestination.Set(childSpan, err.Error())
				childSpan.LogKV("event", "error", "message", err.Error())
			} else {
				ext.Component.Set(childSpan, resp.MessageId)
			}
			childSpan.Finish()
		}()

		tracer.CustomTag("mq.service", "RocketMQ").Set(childSpan)
		tracer.CustomTag("mq.topic", topicName).Set(childSpan)
		tracer.CustomTag("mq.msg_tag", msg.MessageTag).Set(childSpan)
		ext.SpanKindProducer.Set(childSpan)
	}

Next:
	resp, err = p.client.Producer(topicName).PublishMessage(msg)
	if err != nil {
		p.logger.Errorf("发布信息失败[%][%s]:[%s]", topicName, err)
		return err
	}
	p.logger.Infof("Publish成功 ---->MessageId:[%s], BodyMD5:[%s];", resp.MessageId, resp.MessageBodyMD5)
	return nil
}
