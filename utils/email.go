package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
)

const SenderEmail = "no-reply@x-meta.com"

type EmailService struct {
	client *sesv2.Client
}

var emailService *EmailService

func InitEmail() error {
	cfg, err := loadAWSConfig(context.TODO())
	if err != nil {
		log.Printf("[Email] SES init failed: %v", err)
		return fmt.Errorf("failed to load AWS config for SES: %w", err)
	}

	emailService = &EmailService{
		client: sesv2.NewFromConfig(cfg),
	}
	log.Printf("[Email] SES initialized region=%s source=%s", cfg.Region, SenderEmail)
	return nil
}

func GetEmailService() *EmailService {
	return emailService
}

func (s *EmailService) SendTemplatedEmail(to string, templateName string, templateData map[string]string) error {
	dataJSON, err := json.Marshal(templateData)
	if err != nil {
		return fmt.Errorf("marshal template data: %w", err)
	}

	log.Printf(
		"[Email] sending templated email to=%s source=%s template=%s dataKeys=%v",
		maskEmailForLog(to),
		SenderEmail,
		templateName,
		templateDataKeys(templateData),
	)

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(SenderEmail),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateName: aws.String(templateName),
				TemplateData: aws.String(string(dataJSON)),
			},
		},
	}

	out, err := s.client.SendEmail(context.TODO(), input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			log.Printf(
				"[Email] SES send failed to=%s template=%s code=%s message=%s fault=%s",
				maskEmailForLog(to),
				templateName,
				apiErr.ErrorCode(),
				apiErr.ErrorMessage(),
				apiErr.ErrorFault(),
			)
		} else {
			log.Printf("[Email] SES send failed to=%s template=%s error=%v", maskEmailForLog(to), templateName, err)
		}
		return fmt.Errorf("SES send email to %s (template=%s): %w", to, templateName, err)
	}

	log.Printf("[Email] sent template=%s to=%s messageId=%s", templateName, maskEmailForLog(to), aws.ToString(out.MessageId))
	return nil
}

func templateDataKeys(data map[string]string) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

func maskEmailForLog(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at <= 1 {
		return email
	}
	local := email[:at]
	if len(local) <= 2 {
		return local[:1] + "***" + email[at:]
	}
	return local[:2] + "***" + email[at:]
}

func resolveTemplateName(base string, locale string) string {
	if locale == "en" {
		return base + "En"
	}
	return base + "Mn"
}

func emailTimezone() string {
	ulaanbaatar := time.FixedZone("UTC+08:00", 8*60*60)
	return time.Now().In(ulaanbaatar).Format("1/2/2006, 3:04:05 PM UTC-07:00")
}

func (s *EmailService) SendPartnerApplicationEmail(email string, locale string) {
	templateName := resolveTemplateName("partnerRequestTemplate", locale)
	if err := s.SendTemplatedEmail(email, templateName, map[string]string{
		"email":    email,
		"timezone": emailTimezone(),
	}); err != nil {
		log.Printf("ERROR: SendPartnerApplicationEmail: %v", err)
	}
}

func (s *EmailService) SendPartnerApprovedEmail(email string, locale string) {
	templateName := resolveTemplateName("partnerRequestApprovedTemplate", locale)
	if err := s.SendTemplatedEmail(email, templateName, map[string]string{
		"email":    email,
		"timezone": emailTimezone(),
	}); err != nil {
		log.Printf("ERROR: SendPartnerApprovedEmail: %v", err)
	}
}
