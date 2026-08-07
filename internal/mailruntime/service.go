package mailruntime

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	stdmail "net/mail"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/wavefnd/wave-platform/internal/config"
	"github.com/wavefnd/wave-platform/internal/identity"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/storage"
)

const maxDeliveryAttempts = 8

type Sender interface {
	Send(delivery maildomain.Delivery, raw []byte) error
}

type Service struct {
	config   config.MailConfig
	identity *identity.Service
	mail     *maildomain.Repository
	sender   Sender
	servers  []*smtp.Server
	mu       sync.Mutex
}

func New(cfg config.MailConfig, database *storage.Database, identityService *identity.Service) (*Service, error) {
	service := &Service{config: cfg, identity: identityService, mail: maildomain.NewRepository(database)}
	sender, err := newSMTPSender(cfg)
	if err != nil {
		return nil, err
	}
	service.sender = sender
	return service, nil
}

func (service *Service) Run(ctx context.Context) error {
	tlsConfig, err := service.tlsConfig()
	if err != nil {
		return err
	}
	result := make(chan error, 1)
	started := 0
	if address := strings.TrimSpace(service.config.SMTPAddress); address != "" {
		listener, listenErr := net.Listen("tcp", address)
		if listenErr != nil {
			return fmt.Errorf("listen for SMTP mail on %s: %w", address, listenErr)
		}
		backend := &smtpBackend{identity: service.identity, maxMessageBytes: service.config.MaxMessageBytes}
		server := smtp.NewServer(backend)
		server.Domain = service.config.Hostname
		server.TLSConfig = tlsConfig
		server.MaxMessageBytes = service.config.MaxMessageBytes
		server.MaxRecipients = 50
		server.AllowInsecureAuth = false
		server.ReadTimeout = 5 * time.Minute
		server.WriteTimeout = 5 * time.Minute
		service.mu.Lock()
		service.servers = append(service.servers, server)
		service.mu.Unlock()
		started++
		log.Printf("Wave Mail SMTP ingress listening on %s", address)
		go func(current *smtp.Server, currentListener net.Listener) {
			if serveErr := current.Serve(currentListener); serveErr != nil && !errors.Is(serveErr, smtp.ErrServerClosed) {
				result <- fmt.Errorf("serve SMTP mail: %w", serveErr)
				return
			}
			result <- nil
		}(server, listener)
	}

	workerContext, stopWorker := context.WithCancel(ctx)
	defer stopWorker()
	go service.runDeliveryWorker(workerContext)
	if started == 0 {
		<-ctx.Done()
		return nil
	}
	select {
	case <-ctx.Done():
		service.shutdown()
		for range started {
			<-result
		}
		return nil
	case serveErr := <-result:
		service.shutdown()
		return serveErr
	}
}

func (service *Service) shutdown() {
	service.mu.Lock()
	servers := append([]*smtp.Server(nil), service.servers...)
	service.mu.Unlock()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, server := range servers {
		_ = server.Shutdown(shutdownContext)
	}
}

func (service *Service) tlsConfig() (*tls.Config, error) {
	if service.config.TLSCertificate == "" {
		return nil, nil
	}
	certificate, err := tls.LoadX509KeyPair(service.config.TLSCertificate, service.config.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("load SMTP TLS certificate: %w", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12}, nil
}

func (service *Service) runDeliveryWorker(ctx context.Context) {
	service.processDeliveries()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			service.processDeliveries()
		}
	}
}

func (service *Service) processDeliveries() {
	now := time.Now().UTC()
	deliveries, err := service.mail.PendingDeliveries(now, 20)
	if err != nil {
		log.Printf("load pending mail deliveries: %v", err)
		return
	}
	for _, delivery := range deliveries {
		message, err := service.mail.Message(delivery.MessageID)
		if err != nil {
			service.deferDelivery(delivery, fmt.Errorf("load message: %w", err))
			continue
		}
		raw, err := service.mail.RawMessage(message)
		if err != nil {
			service.deferDelivery(delivery, fmt.Errorf("load RFC message: %w", err))
			continue
		}
		delivery.Status = "delivering"
		delivery.Attempts++
		delivery.LastAttemptAt = now
		delivery.LastError = ""
		if err := service.mail.UpsertDelivery(delivery); err != nil {
			log.Printf("mark mail delivery %s active: %v", delivery.ID, err)
			continue
		}
		if err := service.sender.Send(delivery, raw); err != nil {
			service.deferDelivery(delivery, err)
			continue
		}
		delivery.Status = "delivered"
		delivery.CompletedAt = time.Now().UTC()
		delivery.NextAttemptAt = time.Time{}
		if err := service.mail.UpsertDelivery(delivery); err != nil {
			log.Printf("complete mail delivery %s: %v", delivery.ID, err)
			continue
		}
		domain := "unknown"
		if parts := strings.Split(delivery.Recipient, "@"); len(parts) == 2 {
			domain = parts[1]
		}
		log.Printf("Mail delivery %s to %s delivered after attempt %d", delivery.ID, domain, delivery.Attempts)
	}
}

func (service *Service) deferDelivery(delivery maildomain.Delivery, deliveryErr error) {
	now := time.Now().UTC()
	if delivery.Attempts >= maxDeliveryAttempts {
		delivery.Status = "failed"
		delivery.CompletedAt = now
		delivery.NextAttemptAt = time.Time{}
	} else {
		delivery.Status = "deferred"
		delay := time.Minute * time.Duration(1<<min(delivery.Attempts, 8))
		delivery.NextAttemptAt = now.Add(delay)
	}
	delivery.LastError = truncateError(deliveryErr)
	if err := service.mail.UpsertDelivery(delivery); err != nil {
		log.Printf("defer mail delivery %s: %v", delivery.ID, err)
		return
	}
	domain := "unknown"
	if parts := strings.Split(delivery.Recipient, "@"); len(parts) == 2 {
		domain = parts[1]
	}
	log.Printf("Mail delivery %s to %s is %s after attempt %d: %s", delivery.ID, domain,
		delivery.Status, delivery.Attempts, delivery.LastError)
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	value := strings.Join(strings.Fields(err.Error()), " ")
	if len(value) > 500 {
		return value[:500]
	}
	return value
}

type smtpBackend struct {
	identity        *identity.Service
	maxMessageBytes int64
}

func (backend *smtpBackend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &smtpSession{backend: backend}, nil
}

type smtpSession struct {
	backend    *smtpBackend
	from       string
	recipients []string
}

func (session *smtpSession) AuthMechanisms() []string {
	return nil
}

func (session *smtpSession) Auth(string) (sasl.Server, error) {
	return nil, smtp.ErrAuthUnsupported
}

func (session *smtpSession) Mail(from string, _ *smtp.MailOptions) error {
	if from != "" {
		if _, err := stdmail.ParseAddress(from); err != nil {
			return &smtp.SMTPError{Code: 553, EnhancedCode: smtp.EnhancedCode{5, 1, 7}, Message: "Invalid sender address"}
		}
	}
	session.from = strings.ToLower(strings.TrimSpace(from))
	return nil
}

func (session *smtpSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	address, err := stdmail.ParseAddress(to)
	if err != nil {
		return &smtp.SMTPError{Code: 553, EnhancedCode: smtp.EnhancedCode{5, 1, 3}, Message: "Invalid recipient address"}
	}
	normalized := strings.ToLower(address.Address)
	if !session.backend.identity.HasLocalRecipient(normalized) {
		return &smtp.SMTPError{Code: 550, EnhancedCode: smtp.EnhancedCode{5, 1, 1}, Message: "Mailbox unavailable"}
	}
	session.recipients = append(session.recipients, normalized)
	return nil
}

func (session *smtpSession) Data(reader io.Reader) error {
	data, err := io.ReadAll(io.LimitReader(reader, session.backend.maxMessageBytes+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > session.backend.maxMessageBytes {
		return smtp.ErrDataTooLarge
	}
	if _, err := session.backend.identity.AcceptSMTP(nil, session.from, session.recipients, data); err != nil {
		if errors.Is(err, identity.ErrRelayDenied) {
			return &smtp.SMTPError{Code: 550, EnhancedCode: smtp.EnhancedCode{5, 7, 1}, Message: "Relay denied"}
		}
		if errors.Is(err, identity.ErrInvalidMail) {
			return &smtp.SMTPError{Code: 554, EnhancedCode: smtp.EnhancedCode{5, 6, 0}, Message: "Invalid message content"}
		}
		return err
	}
	return nil
}

func (session *smtpSession) Reset() {
	session.from = ""
	session.recipients = nil
}

func (session *smtpSession) Logout() error { return nil }

type smtpSender struct {
	config config.MailConfig
	signer crypto.Signer
}

func newSMTPSender(cfg config.MailConfig) (*smtpSender, error) {
	sender := &smtpSender{config: cfg}
	if cfg.DKIMPrivateKey == "" {
		return sender, nil
	}
	data, err := os.ReadFile(cfg.DKIMPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("read DKIM private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("decode DKIM private key: PEM data is missing")
	}
	var key any
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}
	if err != nil {
		return nil, fmt.Errorf("parse DKIM private key: %w", err)
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("DKIM private key does not implement crypto.Signer")
	}
	if rsaKey, ok := signer.(*rsa.PrivateKey); ok && rsaKey.N.BitLen() < 2048 {
		return nil, errors.New("DKIM RSA private key must be at least 2048 bits")
	}
	sender.signer = signer
	return sender, nil
}

func (sender *smtpSender) Send(delivery maildomain.Delivery, raw []byte) error {
	var err error
	raw, err = sender.sign(raw)
	if err != nil {
		return err
	}
	if sender.config.RelayAddress != "" {
		var auth sasl.Client
		if sender.config.RelayUsername != "" {
			auth = sasl.NewPlainClient("", sender.config.RelayUsername, sender.config.RelayPassword)
		}
		if sender.config.RelayImplicitTLS {
			return smtp.SendMailTLS(sender.config.RelayAddress, auth, delivery.Sender, []string{delivery.Recipient}, bytes.NewReader(raw))
		}
		return smtp.SendMail(sender.config.RelayAddress, auth, delivery.Sender, []string{delivery.Recipient}, bytes.NewReader(raw))
	}
	if !sender.config.DirectDelivery {
		return errors.New("no outbound SMTP relay is configured and direct delivery is disabled")
	}
	address, err := stdmail.ParseAddress(delivery.Recipient)
	if err != nil {
		return err
	}
	parts := strings.Split(address.Address, "@")
	if len(parts) != 2 {
		return errors.New("recipient domain is missing")
	}
	mxRecords, err := net.LookupMX(parts[1])
	if err != nil {
		return fmt.Errorf("lookup MX for %s: %w", parts[1], err)
	}
	var deliveryErrors []error
	for _, record := range mxRecords {
		host := strings.TrimSuffix(record.Host, ".")
		remote := net.JoinHostPort(host, "25")
		client, dialErr := smtp.DialStartTLS(remote, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if dialErr != nil {
			client, dialErr = smtp.Dial(remote)
		}
		if dialErr != nil {
			deliveryErrors = append(deliveryErrors, fmt.Errorf("%s: %w", host, dialErr))
			continue
		}
		if helloErr := client.Hello(sender.config.Hostname); helloErr != nil {
			_ = client.Close()
			deliveryErrors = append(deliveryErrors, fmt.Errorf("%s: %w", host, helloErr))
			continue
		}
		sendErr := client.SendMail(delivery.Sender, []string{delivery.Recipient}, bytes.NewReader(raw))
		if sendErr == nil {
			sendErr = client.Quit()
		} else {
			_ = client.Close()
		}
		if sendErr == nil {
			return nil
		}
		deliveryErrors = append(deliveryErrors, fmt.Errorf("%s: %w", host, sendErr))
	}
	if len(deliveryErrors) == 0 {
		return errors.New("recipient domain has no MX servers")
	}
	return errors.Join(deliveryErrors...)
}

func (sender *smtpSender) sign(raw []byte) ([]byte, error) {
	if sender.signer == nil {
		return raw, nil
	}
	var signed bytes.Buffer
	if err := dkim.Sign(&signed, bytes.NewReader(raw), &dkim.SignOptions{
		Domain: sender.config.DKIMDomain, Selector: sender.config.DKIMSelector, Signer: sender.signer,
		HeaderCanonicalization: dkim.CanonicalizationRelaxed,
		BodyCanonicalization:   dkim.CanonicalizationRelaxed,
		HeaderKeys:             []string{"From", "To", "Cc", "Subject", "Date", "Message-ID", "MIME-Version", "Auto-Submitted", "X-Auto-Response-Suppress", "Content-Type", "Content-Transfer-Encoding"},
	}); err != nil {
		return nil, fmt.Errorf("sign message with DKIM: %w", err)
	}
	return signed.Bytes(), nil
}
