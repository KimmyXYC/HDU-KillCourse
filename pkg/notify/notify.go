package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/cr4n5/HDU-KillCourse/config"
	"github.com/cr4n5/HDU-KillCourse/log"
	"github.com/cr4n5/HDU-KillCourse/util"
)

type TemplateData struct {
	Title string
	Body  string
	Time  string
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Notify 按启用的通道发送通知；任何通道失败都会记录日志，但不会中断其他通道。
func Notify(cfg *config.Config, title string, body string) {
	if cfg == nil {
		return
	}

	// 邮件
	if cfg.SmtpEmail.Enabled == "1" {
		if err := util.SendEmail(cfg.SmtpEmail.Host, cfg.SmtpEmail.Username, cfg.SmtpEmail.Password, cfg.SmtpEmail.To, title, body); err != nil {
			log.Error("发送邮件失败: ", err)
		}
	}

	// Telegram
	if cfg.Telegram.Enabled == "1" {
		if err := sendTelegram(cfg.Telegram, title, body); err != nil {
			log.Error("发送 Telegram 失败: ", err)
		}
	}

	// Bark
	if cfg.Bark.Enabled == "1" {
		if err := sendBark(cfg.Bark, title, body); err != nil {
			log.Error("发送 Bark 失败: ", err)
		}
	}

	// Webhook
	if cfg.Webhook.Enabled == "1" {
		if err := sendWebhook(cfg.Webhook, title, body); err != nil {
			log.Error("发送 Webhook 失败: ", err)
		}
	}
}

func sendTelegram(cfg config.Telegram, title string, body string) error {
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return errors.New("telegram 配置缺失")
	}
	apiBase := strings.TrimRight(cfg.ApiBase, "/")
	if apiBase == "" {
		apiBase = "https://api.telegram.org"
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", apiBase, cfg.BotToken)
	payload := map[string]any{
		"chat_id": cfg.ChatID,
		"text":    fmt.Sprintf("%s\n\n%s", title, body),
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram http %d: %s", resp.StatusCode, string(data))
	}
	return nil
}

func sendBark(cfg config.Bark, title string, body string) error {
	if cfg.Key == "" {
		return errors.New("bark 配置缺失")
	}
	server := strings.TrimRight(cfg.Server, "/")
	if server == "" {
		server = "https://api.day.app"
	}

	// Bark 走 URL Path 形式：/KEY/TITLE/BODY?query...
	// 注意 title/body 需要 path escape，否则包含斜杠会被拆段
	pushURL := fmt.Sprintf("%s/%s/%s/%s", server, url.PathEscape(cfg.Key), url.PathEscape(title), url.PathEscape(body))
	q := url.Values{}
	if cfg.Sound != "" {
		q.Set("sound", cfg.Sound)
	}
	if cfg.Group != "" {
		q.Set("group", cfg.Group)
	}
	if cfg.URL != "" {
		q.Set("url", cfg.URL)
	}
	if cfg.Icon != "" {
		q.Set("icon", cfg.Icon)
	}
	if encoded := q.Encode(); encoded != "" {
		pushURL = pushURL + "?" + encoded
	}

	req, err := http.NewRequest(http.MethodGet, pushURL, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bark http %d: %s", resp.StatusCode, string(data))
	}
	return nil
}

func sendWebhook(cfg config.Webhook, title string, body string) error {
	if cfg.Url == "" {
		return errors.New("webhook 配置缺失")
	}
	method := strings.TrimSpace(cfg.Method)
	if method == "" {
		method = http.MethodPost
	}

	bodyTemplate := strings.TrimSpace(cfg.BodyTemplate)
	if bodyTemplate == "" {
		bodyTemplate = `{"title":"{{.Title}}","body":"{{.Body}}"}`
	}
	data := TemplateData{
		Title: title,
		Body:  body,
		Time:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tpl, err := template.New("webhook").Parse(bodyTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	req, err := http.NewRequest(method, cfg.Url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	if cfg.Headers != nil {
		for k, v := range cfg.Headers {
			if strings.TrimSpace(k) == "" {
				continue
			}
			req.Header.Set(k, v)
		}
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook http %d: %s", resp.StatusCode, string(data))
	}
	return nil
}
