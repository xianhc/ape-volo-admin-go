package email

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wneessen/go-mail"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"go.uber.org/zap"
)

// sendTimeout 是单封邮件的发送超时，避免 SMTP 服务器无响应时长期占用 goroutine。
const sendTimeout = 30 * time.Second

// SendEmail
// @description: 发送邮件
// @param: from 发件人
// @param: fromName 发件人名称
// @param: password 密码
// @param: host 服务器地址
// @param: port 端口
// @param: enableSsl 是否使用 SSL（隐式 TLS，通常是 465 端口）
// @param: to 收件人
// @param: cc 抄送人
// @param: bcc 加密抄送人
// @param: subject 主题
// @param: body 内容（HTML）
// @return: error
func SendEmail(ctx context.Context, from string, fromName string, password string, host string, port int32, enableSsl bool, to []string, cc []string, bcc []string, subject string, body string) error {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	if len(to) == 0 {
		return errors.New("收件人不能为空")
	}

	msg := mail.NewMsg()
	if fromName != "" {
		if err := msg.FromFormat(fromName, from); err != nil {
			return fmt.Errorf("设置发件人失败: %w", err)
		}
	} else if err := msg.From(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	if err := msg.To(to...); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}
	if len(cc) > 0 {
		if err := msg.Cc(cc...); err != nil {
			return fmt.Errorf("设置抄送人失败: %w", err)
		}
	}
	if len(bcc) > 0 {
		if err := msg.Bcc(bcc...); err != nil {
			return fmt.Errorf("设置密送人失败: %w", err)
		}
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	opts := []mail.Option{
		mail.WithPort(int(port)),
		mail.WithTimeout(sendTimeout),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(from),
		mail.WithPassword(password),
	}
	if enableSsl {
		// 隐式 TLS（SMTPS）
		opts = append(opts, mail.WithSSL())
	} else {
		// 服务端支持 STARTTLS 时自动升级，否则使用明文连接。
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		global.Logger.Error("创建邮件客户端失败", zap.String("host", host), zap.Error(err))
		return fmt.Errorf("创建邮件客户端失败: %w", err)
	}

	if err = client.DialAndSendWithContext(ctx, msg); err != nil {
		global.Logger.Error("发送邮件失败", zap.String("host", host), zap.Error(err))
		return err
	}
	return nil
}
