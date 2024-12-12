package email

import (
	"crypto/tls"
	"fmt"
	"github.com/jordan-wright/email"
	"go-apevolo/global"
	"go.uber.org/zap"
	"net/smtp"
)

// SendEmail
// @description: 发送邮件
// @param: from 发件人
// @param: fromName 发件人名称
// @param: password 密码
// @param: host 服务器地址
// @param: port 端口
// @param: enableSsl 示范ssl
// @param: to 收件人
// @param: cc 抄送人
// @param: bcc 加密抄送人
// @param: subject 主题
// @param: body 内容
// @return: error
func SendEmail(from string, fromName string, password string, host string, port int32, enableSsl bool, to []string, cc []string, bcc []string, subject string, body string) error {
	auth := smtp.PlainAuth("", from, password, host)
	e := email.NewEmail()
	if fromName != "" {
		e.From = fmt.Sprintf("%s <%s>", fromName, from)
	} else {
		e.From = from
	}
	e.To = to
	e.Cc = cc
	e.Bcc = bcc
	e.Subject = subject
	e.HTML = []byte(body)
	var err error
	hostAddr := fmt.Sprintf("%s:%d", host, port)
	if enableSsl {
		err = e.SendWithTLS(hostAddr, auth, &tls.Config{ServerName: host})
	} else {
		err = e.Send(hostAddr, auth)
	}
	if err != nil {
		global.Logger.Error("send email error: ", zap.Error(err))
	}
	return err
}
