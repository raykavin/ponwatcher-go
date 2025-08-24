package unm

import (
	"context"
	"errors"
	"fmt"
	"pon_watcher/pkg/log"
	"regexp"
	"strings"
	"sync"
)

const (
	// Response patterns
	ErrorPattern    = "EADD=(.*)"
	HeaderLines     = 8
	FooterLines     = -2
	RequiredColumns = 13

	// Command templates
	LoginCommand   = "LOGIN:::CTAG::UN=%s,PWD=%s;"
	LogoutCommand  = "LOGOUT:::CTAG::;"
	OnuListCommand = "LST-ONU::OLTID=%s,PONID=NA-NA-%d-%d:CTAG::;"
	OnuInfoCommand = "LST-OMDDM::OLTID=%s,PONID=NA-NA-%d-%d,ONUIDTYPE=MAC,ONUID=%s:CTAG::;"

	// Retry configuration
	MaxRetryAttempts = 3
)

var (
	ErrEmptyHostOrPort    = errors.New("host or port cannot be empty")
	ErrConnectionNotFound = errors.New("connection not established")
	ErrInvalidResponse    = errors.New("invalid response format")
	ErrInsufficientData   = errors.New("insufficient data in response")
	ErrInvalidFormat      = errors.New("invalid response format")
	ErrIllegalSession     = errors.New("illegal session")
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
)

// Transporter is a interface to transport TCP connection commands
type Transporter interface {
	Close() error
	Reconnect() error
	IsConnected() bool
	Send(ctx context.Context, cmd string) (string, error)
}

// UNMServer represents a connection to a UNM server with improved error handling and logging.
type UNMServer struct {
	username    string
	password    string
	transporter Transporter
	log         log.Smart
	mtx         sync.Mutex
	connected   bool

	errorRegex *regexp.Regexp
}

// New creates a new UNMServer instance with pre-compiled regex patterns.
func New(username, password string, transporter Transporter, log log.Smart) *UNMServer {
	return &UNMServer{
		username:    username,
		password:    password,
		transporter: transporter,
		log:         log,
		errorRegex:  regexp.MustCompile(ErrorPattern),
	}
}

// isIllegalSessionError checks if the error is of type "illegal session"
func (us *UNMServer) isIllegalSessionError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "illegal session")
}

// executeWithRetry executes a function with automatic retry in case of illegal session
func (us *UNMServer) executeWithRetry(ctx context.Context, operation func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt < MaxRetryAttempts; attempt++ {
		// Ensure connection exists before executing the operation
		if err := us.ensureConnection(ctx); err != nil {
			lastErr = err
			us.log.WithError(err).WithField("attempt", attempt+1).Warn("Failed to ensure connection")
			continue
		}

		// Execute the operation
		err := operation(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// If it's an illegal session error, mark as disconnected and retry
		if us.isIllegalSessionError(err) {
			us.log.WithError(err).WithField("attempt", attempt+1).Warn("Illegal session detected, marking as disconnected")
			us.mtx.Lock()
			us.connected = false
			us.mtx.Unlock()

			// If it's not the last attempt, continue the loop
			if attempt < MaxRetryAttempts-1 {
				continue
			}
		} else {
			// If not a session error, don't retry
			return err
		}
	}

	return fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
}

// shakeHand sends a handshake to maintain the keep alive connection
func (us *UNMServer) shakeHand(ctx context.Context) error {
	response, err := us.sendCommand(ctx, "SHAKEHAND:::CTAG::;")
	if err != nil {
		return err
	}

	if err := us.isResponseErr(response); err != nil {
		return err
	}

	return nil
}

// sendCommand is a helper method that combines command sending and error parsing
func (us *UNMServer) sendCommand(ctx context.Context, command string) (string, error) {
	response, err := us.transporter.Send(ctx, command)
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	if err := us.isResponseErr(response); err != nil {
		return "", err
	}

	return response, nil
}

// ensureConnection sends a handshake to check if there's an active connection
// if the connection is down, it reconnects
func (us *UNMServer) ensureConnection(ctx context.Context) error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	// If already marked as connected, test the connection
	if us.connected {
		/*
		 * OBS:
		 * SENDING A HANDSHAKE TO TEST THE
		 * CONNECTION CAUSES DEGRADED PERFORMANCE
		 *
		 * // Test existing connection with handshake
		 * if err := us.shakeHand(ctx); err != nil {
		 * 	us.log.WithError(err).Debug("Handshake failed, marking as disconnected")
		 * 	us.connected = false
		 * } else {
		 * 	return nil
		 * }
		 */

		return nil
	}

	// First check if we have a connection at transport level
	if !us.transporter.IsConnected() {
		if err := us.reconnectAndLogin(ctx); err != nil {
			return fmt.Errorf("failed to establish connection: %w", err)
		}
		return nil
	}

	// If we got here, we need to reconnect
	us.log.Debug("Attempting reconnection")

	// Close existing connection without holding the mutex
	us.close()

	// Reconnect and login
	if err := us.reconnectAndLogin(ctx); err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	us.connected = true
	return nil
}

// reconnectAndLogin handles the reconnection and login process
func (us *UNMServer) reconnectAndLogin(ctx context.Context) error {
	if err := us.transporter.Reconnect(); err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}

	if err := us.Login(ctx); err != nil {
		us.transporter.Close()
		return fmt.Errorf("login after reconnect failed: %w", err)
	}

	return nil
}

// isResponseErr extracts error information from UNM server response.
func (us *UNMServer) isResponseErr(response string) error {
	if matches := us.errorRegex.FindStringSubmatch(response); len(matches) > 1 {
		errorMsg := strings.TrimSpace(matches[1])
		if errorMsg != "" {
			return fmt.Errorf("UNM server error: %s", errorMsg)
		}
	}

	return nil
}

// Login authenticates with the UNM server.
func (us *UNMServer) Login(ctx context.Context) error {
	command := fmt.Sprintf(LoginCommand, us.username, us.password)

	if _, err := us.sendCommand(ctx, command); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	us.log.Debug("Successfully logged in to UNM server")
	return nil
}

// Logout logs out from the UNM server.
func (us *UNMServer) Logout(ctx context.Context) error {
	if !us.transporter.IsConnected() {
		return nil // Already disconnected
	}

	if _, err := us.sendCommand(ctx, LogoutCommand); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	us.log.Debug("Successfully logged out from UNM server")
	return nil
}

// close performs cleanup and close connection
func (us *UNMServer) close() error {
	us.connected = false

	var errs []error

	// Attempt logout first
	if err := us.Logout(context.Background()); err != nil {
		errs = append(errs, err)
	}

	// Close connection if it exists
	if err := us.transporter.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	us.log.Debug("Successfully closed UNM server and transport connection")
	return nil
}

// Close gracefully closes the connection to the UNM server.
func (us *UNMServer) Close() error {
	us.mtx.Lock()
	defer us.mtx.Unlock()

	return us.close()
}

// FindAllOpticalNetworkUnits queries all ONUs connected to a specified PON port.
func (us *UNMServer) FindAllOpticalNetworkUnits(
	ctx context.Context,
	olt string,
	ponSlot, ponNumber uint,
	filter string,
) ([]*OpticalNetworkUnit, error) {
	var result []*OpticalNetworkUnit

	err := us.executeWithRetry(ctx, func(ctx context.Context) error {
		command := fmt.Sprintf(OnuListCommand, olt, ponSlot, ponNumber)

		response, err := us.sendCommand(ctx, command)
		if err != nil {
			return fmt.Errorf("failed to query ONUs: %w", err)
		}

		onus, err := us.buildONUsFromResponse(response, filter)
		if err != nil {
			return fmt.Errorf("failed to parse ONU response: %w", err)
		}

		result = onus
		return nil
	})

	if err != nil {
		return nil, err
	}

	us.log.WithField("onu_count", len(result)).Debug("Successfully retrieved ONUs")
	return result, nil
}

func (us *UNMServer) FetchAllOpticalNetworkUnitInformation(
	ctx context.Context,
	ponSlot, ponNumber uint,
	olt, physicalAddr string,
) (*OpticalNetworkUnitInfo, error) {
	var result *OpticalNetworkUnitInfo

	err := us.executeWithRetry(ctx, func(ctx context.Context) error {
		command := fmt.Sprintf(OnuInfoCommand, olt, ponSlot, ponNumber, physicalAddr)

		response, err := us.sendCommand(ctx, command)
		if err != nil {
			return fmt.Errorf("failed to query ONU information: %w", err)
		}

		onuInfo, err := us.buildONUInfoFromResponse(response)
		if err != nil {
			return fmt.Errorf("failed to parse ONU response information: %w", err)
		}

		result = onuInfo
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// parseResponseLines is a common helper for parsing response lines
func (us *UNMServer) parseResponseLines(response string, minLines int) ([]string, error) {
	formattedResult := strings.ReplaceAll(response, "\r", "")
	lines := splitAndTrimLines(formattedResult)

	if len(lines) <= minLines {
		return nil, ErrInsufficientData
	}

	return lines, nil
}

// buildONUInfoFromResponse parses the ONU information response from the server.
func (us *UNMServer) buildONUInfoFromResponse(response string) (*OpticalNetworkUnitInfo, error) {
	lines, err := us.parseResponseLines(response, HeaderLines)
	if err != nil {
		return nil, fmt.Errorf("optical_info received invalid arguments: %w", err)
	}

	resultLine := lines[HeaderLines : len(lines)+FooterLines]
	if len(resultLine) == 0 {
		return nil, ErrInsufficientData
	}

	items := strings.Split(resultLine[0], "\t")
	if len(items) < RequiredColumns {
		return nil, fmt.Errorf("optical_info command result read buffer does not match: expected %d columns, got %d", RequiredColumns, len(items))
	}

	return &OpticalNetworkUnitInfo{
		OnuID:             items[0],
		RxPower:           items[1],
		RxPowerStatus:     items[2],
		TxPower:           items[3],
		TxPowerStatus:     items[4],
		CurrTxBias:        items[5],
		CurrTxBiasStatus:  items[6],
		Temperature:       items[7],
		TemperatureStatus: items[8],
		Voltage:           items[9],
		VoltageStatus:     items[10],
		PTxPower:          items[11],
		PRxPower:          items[12],
	}, nil
}

// splitAndTrimLines extracts non-empty, trimmed lines from input
func splitAndTrimLines(input string) []string {
	lines := strings.Split(input, "\n")
	nonEmptyLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if trimmedLine := strings.TrimSpace(line); trimmedLine != "" {
			nonEmptyLines = append(nonEmptyLines, trimmedLine)
		}
	}

	return nonEmptyLines
}

// buildONUsFromResponse parses the ONU list response from the server.
func (us *UNMServer) buildONUsFromResponse(response string, filter string) ([]*OpticalNetworkUnit, error) {
	lines := strings.Split(response, "\n")
	onus := make([]*OpticalNetworkUnit, 0, len(lines))

	for lineNum, line := range lines {
		if trimmedLine := strings.TrimSpace(line); trimmedLine != "" {
			if onu, err := us.processONULine(trimmedLine, filter); err != nil {
				us.log.WithError(err).WithField("line_number", lineNum).Warn("Failed to parse ONU line")
			} else if onu != nil {
				onus = append(onus, onu)
			}
		}
	}

	return onus, nil
}

// processONULine processes a single ONU line and applies filtering
func (us *UNMServer) processONULine(line, filter string) (*OpticalNetworkUnit, error) {
	attrs := strings.Split(line, "\t")
	if len(attrs) != 12 || attrs[0] == "OLTID" {
		return nil, nil // Skip invalid or header lines
	}

	if filter != "" && !us.matchesFilter(attrs[3], attrs[4], filter) {
		return nil, nil // Skip filtered out items
	}

	return us.createONUFromAttributes(attrs)
}

// matchesFilter checks if name or description matches the filter
func (us *UNMServer) matchesFilter(name, desc, filter string) bool {
	lowerFilter := strings.ToLower(filter)
	lowerName := strings.ToLower(name)
	lowerDesc := strings.ToLower(desc)

	return strings.Contains(lowerName, lowerFilter) || strings.Contains(lowerDesc, lowerFilter)
}

// createONUFromAttributes creates an OpticalNetworkUnit from parsed attributes.
func (us *UNMServer) createONUFromAttributes(attrs []string) (*OpticalNetworkUnit, error) {
	if len(attrs) < 12 {
		return nil, fmt.Errorf("insufficient attributes: expected 12, got %d", len(attrs))
	}

	// Helper function to trim all attributes
	trimmedAttrs := make([]string, len(attrs))
	for i, attr := range attrs {
		trimmedAttrs[i] = strings.TrimSpace(attr)
	}

	return &OpticalNetworkUnit{
		OltID:    trimmedAttrs[0],
		PonID:    trimmedAttrs[1],
		OnuNo:    trimmedAttrs[2],
		Name:     trimmedAttrs[3],
		Desc:     trimmedAttrs[4],
		OnuType:  trimmedAttrs[5],
		IP:       trimmedAttrs[6],
		AuthType: trimmedAttrs[7],
		Mac:      trimmedAttrs[8],
		LoID:     trimmedAttrs[9],
		Pwd:      trimmedAttrs[10],
		SwVer:    trimmedAttrs[11],
	}, nil
}
