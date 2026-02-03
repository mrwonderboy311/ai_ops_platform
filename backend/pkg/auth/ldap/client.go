// Package ldap provides LDAP authentication functionality
package ldap

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-ldap/ldap/v3"
)

// Config represents LDAP configuration
type Config struct {
	// LDAP server URL (e.g., ldap://localhost:389 or ldaps://localhost:636)
	URL string
	// Bind DN for authentication (e.g., cn=admin,dc=example,dc=com)
	BindDN string
	// Bind password for authentication
	BindPassword string
	// Base DN for user search (e.g., ou=users,dc=example,dc=com)
	BaseDN string
	// Search filter for users (e.g., (uid=%s) or (sAMAccountName=%s))
	SearchFilter string
	// User attributes to retrieve
	UserAttributes []string
	// Use TLS (for ldaps:// or StartTLS)
	UseTLS bool
	// Skip TLS certificate verification
	InsecureSkipVerify bool
	// Connection timeout
	Timeout time.Duration
}

// DefaultConfig returns default LDAP configuration
func DefaultConfig() Config {
	return Config{
		URL:               "ldap://localhost:389",
		BindDN:            "",
		BindPassword:      "",
		BaseDN:            "dc=example,dc=com",
		SearchFilter:      "(uid=%s)",
		UserAttributes:    []string{"uid", "cn", "mail", "displayName"},
		UseTLS:            false,
		InsecureSkipVerify: false,
		Timeout:           10 * time.Second,
	}
}

// Client represents an LDAP client
type Client struct {
	config Config
}

// NewClient creates a new LDAP client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
	}
}

// Connect connects to the LDAP server
func (c *Client) Connect() (*ldap.Conn, error) {
	var conn *ldap.Conn
	var err error

	if c.config.UseTLS {
		// Use LDAPS (LDAP over SSL/TLS)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.config.InsecureSkipVerify,
		}
		conn, err = ldap.DialTLS("tcp", c.getLDAPAddress(), tlsConfig)
	} else {
		// Use plain LDAP
		conn, err = ldap.Dial("tcp", c.getLDAPAddress())
		if err == nil && !c.config.UseTLS {
			// Try StartTLS if not using LDAPS
			err = conn.StartTLS(&tls.Config{
				InsecureSkipVerify: c.config.InsecureSkipVerify,
			})
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	// Set timeout
	conn.SetTimeout(c.config.Timeout)

	return conn, nil
}

// Bind performs a bind operation with the configured credentials
func (c *Client) Bind(conn *ldap.Conn) error {
	if c.config.BindDN == "" {
		// Anonymous bind
		return conn.UnauthenticatedBind("")
	}
	return conn.Bind(c.config.BindDN, c.config.BindPassword)
}

// Authenticate authenticates a user with username and password
func (c *Client) Authenticate(username, password string) (*ldap.Entry, error) {
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// First, bind with service account
	if err := c.Bind(conn); err != nil {
		return nil, fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	// Search for the user
	searchRequest := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(c.config.SearchFilter, username),
		c.config.UserAttributes,
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("multiple users found")
	}

	userDN := sr.Entries[0].DN

	// Now bind as the user to verify password
	if err := conn.Bind(userDN, password); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	return sr.Entries[0], nil
}

// GetUserAttribute retrieves a specific attribute from an LDAP entry
func GetUserAttribute(entry *ldap.Entry, attributeName string) string {
	for _, attr := range entry.Attributes {
		if attr.Name == attributeName && len(attr.Values) > 0 {
			return attr.Values[0]
		}
	}
	return ""
}

// getLDAPAddress extracts the host:port from the LDAP URL
func (c *Client) getLDAPAddress() string {
	// Remove ldap:// or ldaps:// prefix
	address := c.config.URL
	if len(address) > 8 && address[:8] == "ldaps://" {
		return address[8:]
	}
	if len(address) > 7 && address[:7] == "ldap://" {
		return address[7:]
	}
	return address
}
