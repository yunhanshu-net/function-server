package caller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/yunhanshu-net/function-server/pkg/interfaces"
)

// QQEmailCaller QQ邮件发送caller
type QQEmailCaller struct {
	// QQ邮箱SMTP配置
	SMTPHost  string `json:"smtp_host"`  // SMTP服务器地址
	SMTPPort  string `json:"smtp_port"`  // SMTP端口
	Username  string `json:"username"`   // QQ邮箱账号
	Password  string `json:"password"`   // QQ邮箱授权码
	FromName  string `json:"from_name"`  // 发件人名称
	FromEmail string `json:"from_email"` // 发件人邮箱
}

// QQEmailRequest QQ邮件发送请求
type QQEmailRequest struct {
	To      []string `json:"to"`      // 收件人邮箱列表
	Subject string   `json:"subject"` // 邮件主题
	Body    string   `json:"body"`    // 邮件正文
	IsHTML  bool     `json:"is_html"` // 是否为HTML格式
}

// QQEmailResponse QQ邮件发送响应
type QQEmailResponse struct {
	Success bool   `json:"success"` // 发送是否成功
	Message string `json:"message"` // 响应消息
	Error   string `json:"error"`   // 错误信息（如果有）
}

// NewQQEmailCaller 创建QQ邮件发送caller
func NewQQEmailCaller() *QQEmailCaller {
	return &QQEmailCaller{
		SMTPHost:  "smtp.qq.com",
		SMTPPort:  "587",
		FromName:  "Function Server",
		FromEmail: "", // 需要配置
	}
}

// Call 实现Caller接口
func (c *QQEmailCaller) Call(req *interfaces.FunctionCallReq) (*interfaces.FunctionCallResp, error) {
	// 解析请求参数
	var emailReq QQEmailRequest
	if err := json.Unmarshal(req.Body, &emailReq); err != nil {
		return c.buildErrorResponse("解析请求参数失败", err), nil
	}

	// 验证必要参数
	if len(emailReq.To) == 0 {
		return c.buildErrorResponse("收件人不能为空", nil), nil
	}
	if emailReq.Subject == "" {
		return c.buildErrorResponse("邮件主题不能为空", nil), nil
	}
	if emailReq.Body == "" {
		return c.buildErrorResponse("邮件正文不能为空", nil), nil
	}

	// 发送邮件
	err := c.sendEmail(emailReq)
	if err != nil {
		return c.buildErrorResponse("邮件发送失败", err), nil
	}

	// 构建成功响应
	response := QQEmailResponse{
		Success: true,
		Message: fmt.Sprintf("邮件发送成功，收件人：%s", strings.Join(emailReq.To, ", ")),
	}

	responseBody, _ := json.Marshal(response)
	return &interfaces.FunctionCallResp{
		Name:   req.Name,
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   responseBody,
	}, nil
}

// sendEmail 发送邮件
func (c *QQEmailCaller) sendEmail(req QQEmailRequest) error {
	// 构建邮件内容
	message := c.buildMessage(req)

	// 配置SMTP认证
	auth := smtp.PlainAuth("", c.Username, c.Password, c.SMTPHost)

	// 配置TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         c.SMTPHost,
	}

	// 建立TLS连接
	conn, err := tls.Dial("tcp", c.SMTPHost+":"+c.SMTPPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("建立TLS连接失败: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, c.SMTPHost)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Quit()

	// 认证
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	// 设置发件人
	if err := client.Mail(c.FromEmail); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	for _, to := range req.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败 %s: %v", to, err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送邮件内容失败: %v", err)
	}
	defer writer.Close()

	_, err = writer.Write(message)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	return nil
}

// buildMessage 构建邮件消息
func (c *QQEmailCaller) buildMessage(req QQEmailRequest) []byte {
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", c.FromName, c.FromEmail)
	headers["To"] = strings.Join(req.To, ", ")
	headers["Subject"] = req.Subject
	headers["MIME-Version"] = "1.0"

	// 根据内容类型设置Content-Type
	if req.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// 构建完整邮件
	var message strings.Builder
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(req.Body)

	return []byte(message.String())
}

// buildErrorResponse 构建错误响应
func (c *QQEmailCaller) buildErrorResponse(message string, err error) *interfaces.FunctionCallResp {
	response := QQEmailResponse{
		Success: false,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}

	responseBody, _ := json.Marshal(response)
	return &interfaces.FunctionCallResp{
		Name:   "qq_email",
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   responseBody,
	}
}

// SetConfig 设置QQ邮箱配置
func (c *QQEmailCaller) SetConfig(username, password, fromEmail, fromName string) {
	c.Username = username
	c.Password = password
	c.FromEmail = fromEmail
	if fromName != "" {
		c.FromName = fromName
	}
}

// GetConfig 获取当前配置
func (c *QQEmailCaller) GetConfig() map[string]string {
	return map[string]string{
		"smtp_host":  c.SMTPHost,
		"smtp_port":  c.SMTPPort,
		"username":   c.Username,
		"from_email": c.FromEmail,
		"from_name":  c.FromName,
	}
}
