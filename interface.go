package gnap

/*
type GrantRequest struct {
	accessToken
	subject
	client
	user
	interact
	capbilities
	existingGrant
}

type AccessTokenRequest struct {
	// required (at least one)
	access []Access
	label *string
	flags []string
}

type ResourceAccess interface {

}

type AccessReference string
*/

type AccessTokenAttribute string
const (
	Bearer AccessTokenAttribute = "bearer"
	Split  AccessTokenAttribute = "split"
)

type StartMode string
const (
	StartRedirect StartMode = "redirect"
	StartApp      StartMode = "app"
	StartUserCode StartMode = "user_code"
)

type FinishMode string
const (
	FinishRedirect FinishMode = "redirect"
	FinishPush FinishMode = "push"
)

type Hint string
const (
	HintUILocales Hint = "ui_locales"
)

type ResourceAccessReference string

type ProofForm string
const(
	DetachedJWS ProofForm = "jwsd"
	AttachedJWS ProofForm = "jws"
	MutualTLS   ProofForm = "mtls"
	Dpop        ProofForm = "dpop"
	HTTPSig     ProofForm = "httpsig"
	OAuthPop    ProofForm = "oauthpop"
)
