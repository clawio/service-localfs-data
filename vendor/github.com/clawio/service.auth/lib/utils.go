package lib

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// idtKey is the context key for an identity.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const idtKey key = 0

// NewContext returns a new Context carrying an Identity pat.
func NewContext(ctx context.Context, idt *Identity) context.Context {
	return context.WithValue(ctx, idtKey, idt)
}

// FromContext extracts the Identity pat from ctx, if present.
func FromContext(ctx context.Context) (*Identity, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	p, ok := ctx.Value(idtKey).(*Identity)
	return p, ok
}

// MustFromContext extracts the identity from ctx.
// If not present it panics.
func MustFromContext(ctx context.Context) *Identity {
	idt, ok := ctx.Value(idtKey).(*Identity)
	if !ok {
		panic("identity is not registered")
	}
	return idt
}

func ParseToken(t, secret string) (*Identity, error) {

	token, err := jwt.Parse(t, func(token *jwt.Token) (key interface{}, err error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	identity := &Identity{}

	pidString, ok := token.Claims["pid"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of pid:%s",
			fmt.Sprintln(token.Claims["pid"]))
	}
	idpString, ok := token.Claims["idp"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of idp:%s",
			fmt.Sprintln(token.Claims["idp"]))
	}
	displaynameString, ok := token.Claims["display_name"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of display_ame:%s",
			fmt.Sprintln(token.Claims["display_ame"]))
	}
	emailString, ok := token.Claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of email:%s",
			fmt.Sprintln(token.Claims["email"]))
	}

	identity.Pid = pidString
	identity.Idp = idpString
	identity.DisplayName = displaynameString
	identity.Email = emailString

	return identity, nil
}

type Identity struct {
	Pid         string `json:"pid"`
	Idp         string `json:"idp"`
	Email       string `json:"email"`
	DisplayName string `json:"email"`
}

func (i *Identity) String() string {
	return fmt.Sprintf("identity(pid:%s idp:%s email:%s)", i.Pid, i.Idp, i.Email)
}
