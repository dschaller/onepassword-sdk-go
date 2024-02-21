package onepassword

import (
	"context"
	"fmt"
	"github.com/1password/onepassword-sdk-go/internal"
	"runtime"
)

const (
	DefaultIntegrationName    = "Unknown"
	DefaultIntegrationVersion = "Unknown"
)

func createClient(core internal.Core, opts ...ClientOption) (*Client, error) {
	client := Client{
		config: internal.NewDefaultConfig(),
	}

	for _, opt := range opts {
		err := opt(&client)
		if err != nil {
			return nil, err
		}
	}

	clientID, err := core.InitClient(client.config)
	if err != nil {
		return nil, fmt.Errorf("error initializing client: %w", err)
	}

	inner := InnerClient{
		id:   *clientID,
		core: core,
	}

	initAPIs(&client, inner)

	runtime.SetFinalizer(&client, func(f *Client) {
		core.ReleaseClient(*clientID)
	})
	return &client, nil
}

// NewClient returns a 1Password Go SDK client using the provided ClientOption list.
func NewClient(sharedContext context.Context, opts ...ClientOption) (*Client, error) {
	core, err := internal.GetSharedCore(sharedContext)
	if err != nil {
		return nil, err
	}
	return createClient(core, opts...)
}

// InnerClient represents the sdk-core client on which calls will be made.
type InnerClient struct {
	id   uint64
	core internal.Core
}

type ClientOption func(client *Client) error

// WithServiceAccountToken specifies the [1Password Service Account](https://developer.1password.com/docs/service-accounts) token to use to authenticate the SDK client.
func WithServiceAccountToken(token string) ClientOption {
	return func(c *Client) error {
		c.config.SAToken = token
		return nil
	}
}

// WithIntegrationInfo specifies the name and version of the integration built using the 1Password Go SDK. If you don't know which name and version to use, use `DefaultIntegrationName` and `DefaultIntegrationVersion`, respectively.
func WithIntegrationInfo(name string, version string) ClientOption {
	return func(c *Client) error {
		c.config.IntegrationName = name
		c.config.IntegrationVersion = version
		return nil
	}
}

func clientInvoke(innerClient InnerClient, invocation string, params map[string]interface{}) (*string, error) {
	return innerClient.core.Invoke(internal.InvokeConfig{
		ClientID: innerClient.id,
		Invocation: internal.Invocation{
			MethodName: invocation,
			Parameters: params,
		},
	})
}
